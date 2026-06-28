package usecase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"llm-gateway/internal/shared/uuid"

	billingUsecase "llm-gateway/internal/billing/usecase"
	channelDomain "llm-gateway/internal/channel/domain"
	"llm-gateway/internal/relay"
	"llm-gateway/internal/relay/adaptor"
	"llm-gateway/internal/shared/config"
	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/errcode"
)

// RelayUsecase 代理引擎核心
type RelayUsecase struct {
	channelSelector *relay.ChannelSelector
	billingUsecase  *billingUsecase.BillingUsecase
	channelRepo     channelDomain.ChannelRepository
	keyCrypto       *crypto.ChaCha20Poly1305Crypto
	httpClient      *http.Client
	httpTransport   *http.Transport
}

// NewRelayUsecase 创建代理引擎
func NewRelayUsecase(
	channelSelector *relay.ChannelSelector,
	billingUsecase *billingUsecase.BillingUsecase,
	channelRepo channelDomain.ChannelRepository,
) *RelayUsecase {
	cfg := config.Load()
	transport := newRelayHTTPTransport()
	return &RelayUsecase{
		channelSelector: channelSelector,
		billingUsecase:  billingUsecase,
		channelRepo:     channelRepo,
		keyCrypto:       crypto.NewChaCha20Poly1305Crypto(cfg.ChaCha20Poly1305Key),
		httpClient:      &http.Client{Transport: transport},
		httpTransport:   transport,
	}
}

func newRelayHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// RelayRequest 代理请求
type RelayRequest struct {
	UserID    string
	KeyID     string
	IPAddress string
	Request   *adaptor.ChatRequest
}

// RelayResponse 代理响应
type RelayResponse struct {
	StatusCode int
	Body       io.Reader
	Headers    map[string]string
}

// ChatCompletion 聊天补全
func (uc *RelayUsecase) ChatCompletion(ctx context.Context, req *RelayRequest) (*RelayResponse, error) {
	// 1. 获取支持该模型的渠道
	channels, err := uc.channelRepo.GetActiveByModelForUser(req.Request.Model, req.UserID)
	if err != nil || len(channels) == 0 {
		return nil, errcode.ErrNoChannel
	}

	// 2. 选择第一个渠道（简化实现，实际应按优先级+权重选择）
	channel := channels[0]

	// 3. 解密API Key
	apiKey, err := uc.keyCrypto.Decrypt(channel.APIKeyEnc)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	// 4. 获取适配器
	ad, err := adaptor.GetAdaptor(channel.Type)
	if err != nil {
		return nil, errcode.ErrNoChannel
	}
	upstreamModel := mapModelName(channel, req.Request.Model)
	upstreamRequest := *req.Request
	upstreamRequest.Model = upstreamModel

	// 5. 预扣费
	preConsumed, err := uc.billingUsecase.PreConsume(ctx, billingUsecase.PreConsumeParams{
		UserID:       req.UserID,
		KeyID:        req.KeyID,
		ChannelID:    channel.ID,
		Model:        req.Request.Model,
		PromptTokens: estimateTokens(req.Request.Messages),
		IsStream:     req.Request.Stream,
		IPAddress:    req.IPAddress,
	})
	if err != nil {
		return nil, err
	}

	// 6. 转换请求格式
	transformedReq, err := ad.TransformRequest(&upstreamRequest)
	if err != nil {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		return nil, errcode.ErrFormatConvert
	}

	// 7. 序列化请求
	reqBody, err := json.Marshal(transformedReq)
	if err != nil {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		return nil, errcode.ErrInternal
	}

	// 8. 构建上游请求
	upstreamURL := ad.GetEndpoint(channel.BaseURL, upstreamModel)
	requestCtx := ctx
	var upstreamCancel context.CancelFunc
	if upstreamRequest.Stream {
		requestCtx, upstreamCancel = context.WithCancel(ctx)
	}

	buildHTTPRequest := func() (*http.Request, error) {
		httpReq, err := http.NewRequestWithContext(requestCtx, "POST", upstreamURL, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		for key, value := range ad.GetHeaders(apiKey) {
			httpReq.Header.Set(key, value)
		}
		return httpReq, nil
	}

	httpReq, err := buildHTTPRequest()
	if err != nil {
		if upstreamCancel != nil {
			upstreamCancel()
		}
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		return nil, errcode.ErrInternal
	}

	// 9. 发送请求
	startTime := time.Now()
	httpResp, err := uc.httpClient.Do(httpReq)
	if err != nil {
		if requestCtx.Err() == nil && isStaleNetworkError(err) {
			uc.httpTransport.CloseIdleConnections()
			httpReq, reqErr := buildHTTPRequest()
			if reqErr == nil {
				httpResp, err = uc.httpClient.Do(httpReq)
			}
		}
	}
	if err != nil {
		if upstreamCancel != nil {
			upstreamCancel()
		}
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		uc.recordChannelStats(channel.ID, false, 0)
		return nil, errcode.ErrUpstreamTimeout
	}
	latency := int(time.Since(startTime).Milliseconds())
	if httpResp.StatusCode >= 400 {
		if upstreamCancel != nil {
			defer upstreamCancel()
		}
		defer httpResp.Body.Close()
		body, _ := io.ReadAll(httpResp.Body)
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		uc.recordChannelStats(channel.ID, false, 0)
		return &RelayResponse{
			StatusCode: httpResp.StatusCode,
			Body:       bytes.NewReader(normalizeErrorBody(body)),
			Headers:    map[string]string{"Content-Type": "application/json", "X-Request-Id": uuid.NewV7String()},
		}, nil
	}

	// 10. 处理响应
	if upstreamRequest.Stream {
		return uc.handleStreamResponse(ctx, httpResp, upstreamCancel, req, channel, ad, preConsumed, latency)
	}
	return uc.handleNonStreamResponse(ctx, httpResp, req, channel, ad, preConsumed, latency)
}

// handleStreamResponse 处理流式响应
func (uc *RelayUsecase) handleStreamResponse(
	ctx context.Context,
	httpResp *http.Response,
	upstreamCancel context.CancelFunc,
	req *RelayRequest,
	channel *channelDomain.Channel,
	ad adaptor.Adaptor,
	preConsumed int64,
	latency int,
) (*RelayResponse, error) {
	headers := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
		"X-Request-Id":  uuid.NewV7String(),
	}
	if adaptor.IsOpenAICompatible(ad) {
		for key, values := range httpResp.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		if headers["Content-Type"] == "" {
			headers["Content-Type"] = "text/event-stream"
		}

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			defer httpResp.Body.Close()
			defer upstreamCancel()

			reader := bufio.NewReader(httpResp.Body)
			promptTokens := estimateTokens(req.Request.Messages)
			completionTokens := 0
			cacheTokens := 0
			contentProduced := false
			watchdogTimeout := relayTTFTTimeout()
			firstMeaningful := make(chan struct{})
			var firstMeaningfulOnce sync.Once
			watchdogTimedOut := startTTFTWatchdog(watchdogTimeout, upstreamCancel, firstMeaningful)
			defer stopTTFTWatchdog(firstMeaningful, &firstMeaningfulOnce)

			for {
				line, err := reader.ReadString('\n')
				if len(line) > 0 {
					trimmed := strings.TrimSpace(line)
					meaningful := isMeaningfulSSEData(trimmed)
					if meaningful {
						stopTTFTWatchdog(firstMeaningful, &firstMeaningfulOnce)
						usage := extractOpenAIUsage([]byte(strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))))
						if usage.PromptTokens > 0 || usage.CompletionTokens > 0 || usage.TotalTokens > 0 {
							promptTokens = usage.PromptTokens
							completionTokens = usage.CompletionTokens
							cacheTokens = usage.CacheTokens
						} else {
							completionTokens++
						}
					}
					if contentProduced || meaningful {
						if _, writeErr := pw.Write([]byte(line)); writeErr != nil {
							break
						}
						contentProduced = true
					}
				}
				if err != nil {
					if !contentProduced && watchdogTimedOut() {
						uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
						uc.recordChannelStats(channel.ID, false, 0)
						return
					}
					if err != io.EOF && !contentProduced {
						uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
						uc.recordChannelStats(channel.ID, false, 0)
						return
					}
					break
				}
			}

			uc.postConsumeAndRecord(ctx, channel.ID, billingUsecase.PostConsumeParams{
				RequestID:        uuid.NewV7String(),
				UserID:           req.UserID,
				KeyID:            req.KeyID,
				ChannelID:        channel.ID,
				Endpoint:         "/v1/chat/completions",
				Model:            req.Request.Model,
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				CacheHit:         cacheTokens > 0,
				CacheTokens:      cacheTokens,
				StatusCode:       httpResp.StatusCode,
				IsStream:         true,
				FirstByteMs:      latency,
				LatencyMs:        latency,
				IPAddress:        req.IPAddress,
				PreConsumed:      preConsumed,
			})
		}()

		return &RelayResponse{StatusCode: httpResp.StatusCode, Body: pr, Headers: headers}, nil
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer httpResp.Body.Close()
		defer upstreamCancel()

		reader := bufio.NewReader(httpResp.Body)
		completionTokens := 0
		contentProduced := false
		watchdogTimeout := relayTTFTTimeout()
		firstMeaningful := make(chan struct{})
		var firstMeaningfulOnce sync.Once
		watchdogTimedOut := startTTFTWatchdog(watchdogTimeout, upstreamCancel, firstMeaningful)
		defer stopTTFTWatchdog(firstMeaningful, &firstMeaningfulOnce)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if !contentProduced && watchdogTimedOut() {
					uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
					uc.recordChannelStats(channel.ID, false, 0)
					return
				}
				if err != io.EOF && !contentProduced {
					uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
					uc.recordChannelStats(channel.ID, false, 0)
					return
				}
				break
			}

			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				if contentProduced {
					_, _ = pw.Write([]byte("data: [DONE]\n\n"))
				}
				break
			}
			if strings.TrimSpace(data) != "" {
				stopTTFTWatchdog(firstMeaningful, &firstMeaningfulOnce)
			}

			chunk, err := ad.TransformStreamChunk([]byte(data))
			if err != nil || chunk == nil {
				continue
			}

			chunkData, _ := json.Marshal(chunk)
			if _, writeErr := pw.Write([]byte("data: " + string(chunkData) + "\n\n")); writeErr != nil {
				break
			}
			contentProduced = true

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				completionTokens++
			}
		}

		uc.postConsumeAndRecord(ctx, channel.ID, billingUsecase.PostConsumeParams{
			RequestID:        uuid.NewV7String(),
			UserID:           req.UserID,
			KeyID:            req.KeyID,
			ChannelID:        channel.ID,
			Endpoint:         "/v1/chat/completions",
			Model:            req.Request.Model,
			PromptTokens:     estimateTokens(req.Request.Messages),
			CompletionTokens: completionTokens,
			StatusCode:       httpResp.StatusCode,
			IsStream:         true,
			FirstByteMs:      latency,
			LatencyMs:        latency,
			IPAddress:        req.IPAddress,
			PreConsumed:      preConsumed,
		})
	}()

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       pr,
		Headers:    headers,
	}, nil
}

// handleNonStreamResponse 处理非流式响应
func (uc *RelayUsecase) handleNonStreamResponse(
	ctx context.Context,
	httpResp *http.Response,
	req *RelayRequest,
	channel *channelDomain.Channel,
	ad adaptor.Adaptor,
	preConsumed int64,
	latency int,
) (*RelayResponse, error) {
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		uc.recordChannelStats(channel.ID, false, 0)
		return nil, errcode.ErrUpstreamInvalid
	}

	if adaptor.IsOpenAICompatible(ad) {
		usage := extractOpenAIUsage(body)
		uc.postConsumeAndRecord(ctx, channel.ID, billingUsecase.PostConsumeParams{
			RequestID:        uuid.NewV7String(),
			UserID:           req.UserID,
			KeyID:            req.KeyID,
			ChannelID:        channel.ID,
			Endpoint:         "/v1/chat/completions",
			Model:            req.Request.Model,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			CacheHit:         usage.CacheTokens > 0,
			CacheTokens:      usage.CacheTokens,
			StatusCode:       httpResp.StatusCode,
			IsStream:         false,
			FirstByteMs:      latency,
			LatencyMs:        latency,
			IPAddress:        req.IPAddress,
			PreConsumed:      preConsumed,
		})

		headers := map[string]string{"Content-Type": "application/json", "X-Request-Id": uuid.NewV7String()}
		if contentType := httpResp.Header.Get("Content-Type"); contentType != "" {
			headers["Content-Type"] = contentType
		}
		return &RelayResponse{StatusCode: httpResp.StatusCode, Body: bytes.NewReader(body), Headers: headers}, nil
	}

	if httpResp.StatusCode >= 400 {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		uc.recordChannelStats(channel.ID, false, 0)
		return &RelayResponse{
			StatusCode: httpResp.StatusCode,
			Body:       bytes.NewReader(normalizeErrorBody(body)),
			Headers:    map[string]string{"Content-Type": "application/json", "X-Request-Id": uuid.NewV7String()},
		}, nil
	}

	resp, err := ad.TransformResponse(body)
	if err != nil {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		uc.recordChannelStats(channel.ID, false, 0)
		return nil, errcode.ErrFormatConvert
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		uc.billingUsecase.Refund(ctx, req.UserID, preConsumed)
		return nil, errcode.ErrInternal
	}

	uc.postConsumeAndRecord(ctx, channel.ID, billingUsecase.PostConsumeParams{
		RequestID:        uuid.NewV7String(),
		UserID:           req.UserID,
		KeyID:            req.KeyID,
		ChannelID:        channel.ID,
		Endpoint:         "/v1/chat/completions",
		Model:            req.Request.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		StatusCode:       httpResp.StatusCode,
		IsStream:         false,
		FirstByteMs:      latency,
		LatencyMs:        latency,
		IPAddress:        req.IPAddress,
		PreConsumed:      preConsumed,
	})

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       bytes.NewReader(respBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Request-Id": uuid.NewV7String(),
		},
	}, nil
}

// estimateTokens 估算token数
func estimateTokens(messages []adaptor.Message) int {
	total := 0
	for _, msg := range messages {
		total += len(messageContentText(msg.Content)) / 4
	}
	return total
}

func mapModelName(channel *channelDomain.Channel, model string) string {
	if channel.ModelMapping == "" {
		return model
	}
	mapping := map[string]string{}
	if err := json.Unmarshal([]byte(channel.ModelMapping), &mapping); err != nil {
		return model
	}
	if mapped, ok := mapping[model]; ok && mapped != "" {
		return mapped
	}
	return model
}

func (uc *RelayUsecase) recordChannelStats(channelID string, success bool, quota int64) {
	if uc.channelRepo == nil || channelID == "" {
		return
	}
	if quota < 0 {
		quota = 0
	}
	_ = uc.channelRepo.UpdateStats(channelID, success, quota)
}

func (uc *RelayUsecase) postConsumeAndRecord(ctx context.Context, channelID string, params billingUsecase.PostConsumeParams) {
	actualCost, err := uc.billingUsecase.PostConsume(ctx, params)
	if err != nil {
		log.Printf("billing post consume failed: request_id=%s channel_id=%s user_id=%s err=%v", params.RequestID, channelID, params.UserID, err)
	}
	uc.recordChannelStats(channelID, true, actualCost)
}

func isStaleNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	msg := strings.ToLower(err.Error())
	phrases := []string{
		"connection reset",
		"broken pipe",
		"server closed idle connection",
		"connection refused",
		"connection is shutting down",
		"use of closed network connection",
		"unexpected eof",
	}
	for _, phrase := range phrases {
		if strings.Contains(msg, phrase) {
			return true
		}
	}
	return msg == "eof"
}

func relayTTFTTimeout() time.Duration {
	value := strings.TrimSpace(os.Getenv("RELAY_TTFT_TIMEOUT_SECONDS"))
	if value == "" {
		return 90 * time.Second
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 90 * time.Second
	}
	if seconds == 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func startTTFTWatchdog(timeout time.Duration, cancel context.CancelFunc, firstMeaningful <-chan struct{}) func() bool {
	var timedOut atomic.Bool
	if timeout <= 0 {
		return timedOut.Load
	}
	go func() {
		select {
		case <-time.After(timeout):
			timedOut.Store(true)
			cancel()
		case <-firstMeaningful:
		}
	}()
	return timedOut.Load
}

func stopTTFTWatchdog(firstMeaningful chan struct{}, once *sync.Once) {
	once.Do(func() { close(firstMeaningful) })
}

func isMeaningfulSSEData(line string) bool {
	if !strings.HasPrefix(line, "data:") {
		return false
	}
	data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	return data != "" && data != "[DONE]"
}

func extractOpenAIUsage(body []byte) adaptor.Usage {
	var resp struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			PromptDetails    struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
			PromptCacheHitTokens int `json:"prompt_cache_hit_tokens"`
		} `json:"usage"`
	}
	_ = json.Unmarshal(body, &resp)
	cacheTokens := resp.Usage.PromptDetails.CachedTokens
	if cacheTokens == 0 {
		cacheTokens = resp.Usage.PromptCacheHitTokens
	}
	return adaptor.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
		CacheTokens:      cacheTokens,
	}
}

func normalizeErrorBody(body []byte) []byte {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err == nil {
		if _, ok := raw["error"]; ok {
			return body
		}
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = "upstream error"
	}
	data, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "upstream_error",
			"param":   nil,
			"code":    "upstream_error",
		},
	})
	return data
}

func messageContentText(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}

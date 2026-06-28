package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	channelDomain "llm-gateway/internal/channel/domain"
	modelDomain "llm-gateway/internal/model/domain"
	"llm-gateway/internal/relay/adaptor"
	"llm-gateway/internal/relay/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
)

type RelayHandler struct {
	relayUsecase *usecase.RelayUsecase
	modelRepo    modelDomain.ModelRepository
	channelRepo  channelDomain.ChannelRepository
}

func NewRelayHandler(relayUsecase *usecase.RelayUsecase, modelRepo modelDomain.ModelRepository, channelRepo channelDomain.ChannelRepository) *RelayHandler {
	return &RelayHandler{relayUsecase: relayUsecase, modelRepo: modelRepo, channelRepo: channelRepo}
}

// ChatCompletion 聊天补全接口
func (h *RelayHandler) ChatCompletion(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	keyID := getKeyID(r.Context())
	ipAddress := r.RemoteAddr

	var req adaptor.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.OpenAIError(w, errcode.ErrInvalidRelayBody)
		return
	}

	if req.Model == "" {
		resp.OpenAIError(w, errcode.ErrModelMissing)
		return
	}
	if len(req.Messages) == 0 {
		resp.OpenAIError(w, errcode.ErrInvalidMessages)
		return
	}

	relayResp, err := h.relayUsecase.ChatCompletion(r.Context(), &usecase.RelayRequest{
		UserID:    userID,
		KeyID:     keyID,
		IPAddress: ipAddress,
		Request:   &req,
	})
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.OpenAIError(w, appErr)
		} else {
			resp.OpenAIError(w, errcode.ErrInternal)
		}
		return
	}

	if req.Stream {
		if _, ok := w.(http.Flusher); !ok {
			resp.OpenAIError(w, errcode.ErrInternal)
			return
		}
	}

	for key, value := range relayResp.Headers {
		w.Header().Set(key, value)
	}
	if req.Stream {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/event-stream")
		}
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("X-Accel-Buffering", "no")
	}
	w.WriteHeader(relayResp.StatusCode)
	if closer, ok := relayResp.Body.(io.Closer); ok {
		defer closer.Close()
	}

	if req.Stream {
		flusher := w.(http.Flusher)
		flusher.Flush()

		buf := make([]byte, 4096)
		for {
			n, readErr := relayResp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					if closer, ok := relayResp.Body.(io.Closer); ok {
						closer.Close()
					}
					return
				}
				flusher.Flush()
			}
			if readErr != nil {
				break
			}
		}
	} else {
		io.Copy(w, relayResp.Body)
	}
}

// GetModels 获取模型列表
func (h *RelayHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		resp.OpenAIError(w, errcode.ErrUnauthorized)
		return
	}

	enabled := true
	channels, err := h.channelRepo.ListActiveForUser(userID)
	if err != nil {
		resp.OpenAIError(w, errcode.ErrDatabase)
		return
	}

	seen := make(map[string]struct{})
	models := make([]map[string]interface{}, 0)
	for _, channel := range channels {
		channelModels, err := h.modelRepo.List(&channel.ID, &enabled, "")
		if err != nil {
			resp.OpenAIError(w, errcode.ErrDatabase)
			return
		}
		for _, model := range channelModels {
			if model.ModelName == "" {
				continue
			}
			if _, exists := seen[model.ModelName]; exists {
				continue
			}
			seen[model.ModelName] = struct{}{}

			models = append(models, map[string]interface{}{
				"id":       model.ModelName,
				"object":   "model",
				"created":  model.CreatedAt.Unix(),
				"owned_by": "system",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}

func getKeyID(ctx context.Context) string {
	if val, ok := ctx.Value("key_id").(string); ok {
		return val
	}
	return ""
}

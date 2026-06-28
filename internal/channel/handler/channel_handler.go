package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/channel/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
)

type ChannelHandler struct {
	channelUsecase *usecase.ChannelUsecase
}

func NewChannelHandler(channelUsecase *usecase.ChannelUsecase) *ChannelHandler {
	return &ChannelHandler{channelUsecase: channelUsecase}
}

func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	var status, channelType *int
	if s := r.URL.Query().Get("status"); s != "" {
		val, _ := strconv.Atoi(s)
		status = &val
	}
	if t := r.URL.Query().Get("type"); t != "" {
		val, _ := strconv.Atoi(t)
		channelType = &val
	}

	channels, total, err := h.channelUsecase.ListChannels(page, size, status, channelType)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Paginated(w, total, page, size, channels)
}

func (h *ChannelHandler) ListModelPlaza(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}
	channels, err := h.channelUsecase.ListModelPlaza(userID)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, channels)
}

func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	channel, err := h.channelUsecase.CreateChannel(req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, channel)
}

func (h *ChannelHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	channel, err := h.channelUsecase.GetChannel(id)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, channel)
}

func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	channel, err := h.channelUsecase.UpdateChannel(id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, channel)
}

func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.channelUsecase.DeleteChannel(id); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "channel deleted successfully"})
}

func (h *ChannelHandler) ToggleChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	channel, err := h.channelUsecase.ToggleChannel(id)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, channel)
}

func (h *ChannelHandler) TestChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	channel, err := h.channelUsecase.GetChannel(id)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	// TODO: 实际测试渠道连接
	resp.Success(w, map[string]interface{}{
		"success":    true,
		"latency_ms": 100,
		"tested_at":  channel.UpdatedAt,
	})
}

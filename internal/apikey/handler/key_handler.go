package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/apikey/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
)

type KeyHandler struct {
	keyUsecase *usecase.KeyUsecase
}

func NewKeyHandler(keyUsecase *usecase.KeyUsecase) *KeyHandler {
	return &KeyHandler{keyUsecase: keyUsecase}
}

func (h *KeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	keys, err := h.keyUsecase.ListKeys(userID)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, keys)
}

func (h *KeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req usecase.CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	result, err := h.keyUsecase.CreateKey(userID, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, result)
}

func (h *KeyHandler) UpdateKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	key, err := h.keyUsecase.UpdateKey(id, userID, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, key)
}

func (h *KeyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.keyUsecase.DeleteKey(id, userID); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			slog.Error("delete api key failed", "key_id", id, "user_id", userID, "error", err)
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "api key deleted successfully"})
}

func (h *KeyHandler) ToggleKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	key, err := h.keyUsecase.ToggleKey(id, userID)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, key)
}

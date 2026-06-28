package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/model/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/resp"
)

type ModelHandler struct {
	modelUsecase *usecase.ModelUsecase
}

func NewModelHandler(modelUsecase *usecase.ModelUsecase) *ModelHandler {
	return &ModelHandler{modelUsecase: modelUsecase}
}

func (h *ModelHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	var channelID *string
	var enabled *bool
	if c := r.URL.Query().Get("channel_id"); c != "" {
		channelID = &c
	}
	if e := r.URL.Query().Get("enabled"); e != "" {
		val := e == "true"
		enabled = &val
	}
	search := r.URL.Query().Get("search")

	models, err := h.modelUsecase.ListModels(channelID, enabled, search)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, models)
}

func (h *ModelHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	model, err := h.modelUsecase.CreateModel(req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, model)
}

func (h *ModelHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.CreateModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	model, err := h.modelUsecase.UpdateModel(id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, model)
}

func (h *ModelHandler) ToggleModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	model, err := h.modelUsecase.ToggleModel(id, req.Enabled)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, model)
}

func (h *ModelHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.modelUsecase.DeleteModel(id); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "model deleted successfully"})
}

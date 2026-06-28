package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/pricing/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/resp"
)

type PricingHandler struct {
	pricingUsecase *usecase.PricingUsecase
}

func NewPricingHandler(pricingUsecase *usecase.PricingUsecase) *PricingHandler {
	return &PricingHandler{pricingUsecase: pricingUsecase}
}

func (h *PricingHandler) ListPricings(w http.ResponseWriter, r *http.Request) {
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

	pricings, err := h.pricingUsecase.ListPricings(channelID, enabled, search)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, pricings)
}

func (h *PricingHandler) CreatePricing(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	pricing, err := h.pricingUsecase.CreatePricing(req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, pricing)
}

func (h *PricingHandler) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.CreatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	pricing, err := h.pricingUsecase.UpdatePricing(id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, pricing)
}

func (h *PricingHandler) TogglePricing(w http.ResponseWriter, r *http.Request) {
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

	pricing, err := h.pricingUsecase.TogglePricing(id, req.Enabled)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, pricing)
}

func (h *PricingHandler) DeletePricing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.pricingUsecase.DeletePricing(id); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "pricing deleted successfully"})
}

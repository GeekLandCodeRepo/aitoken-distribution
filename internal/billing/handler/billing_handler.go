package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/billing/domain"
	"llm-gateway/internal/billing/usecase"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
)

type BillingHandler struct {
	redeemUsecase *usecase.RedeemUsecase
	txRepo        domain.TransactionRepository
}

func NewBillingHandler(redeemUsecase *usecase.RedeemUsecase, txRepo domain.TransactionRepository) *BillingHandler {
	return &BillingHandler{
		redeemUsecase: redeemUsecase,
		txRepo:        txRepo,
	}
}

func (h *BillingHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	result, err := h.redeemUsecase.Redeem(userID, req.Code)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, result)
}

func (h *BillingHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var txType *int
	if t := r.URL.Query().Get("type"); t != "" {
		val, _ := strconv.Atoi(t)
		txType = &val
	}

	txs, total, err := h.txRepo.ListByUserID(userID, page, size, txType)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Paginated(w, total, page, size, txs)
}

func (h *BillingHandler) GetAdminTransactions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	userID := r.URL.Query().Get("user_id")
	userEmail := r.URL.Query().Get("user_email")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var txType *int
	if t := r.URL.Query().Get("type"); t != "" {
		val, _ := strconv.Atoi(t)
		txType = &val
	}

	txs, total, err := h.txRepo.ListAll(page, size, userID, userEmail, txType)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Paginated(w, total, page, size, txs)
}

func (h *BillingHandler) ListCodes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	status := r.URL.Query().Get("status")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	codes, total, err := h.redeemUsecase.ListCodes(page, size, status)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Paginated(w, total, page, size, codes)
}

func (h *BillingHandler) GenerateCodes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Quota     int64  `json:"quota"`
		Count     int    `json:"count"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	codes, err := h.redeemUsecase.GenerateCodes(userID, req.Quota, req.Count, expiresAt)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, map[string]interface{}{
		"count": len(codes),
		"codes": codes,
	})
}

func (h *BillingHandler) DeleteCode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.redeemUsecase.DeleteCode(id); err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, map[string]string{"message": "redeem code deleted successfully"})
}

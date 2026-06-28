package handler

import (
	"net/http"
	"strconv"
	"time"

	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
	"llm-gateway/internal/usage/usecase"
)

type UsageHandler struct {
	usageUsecase *usecase.UsageUsecase
}

func NewUsageHandler(usageUsecase *usecase.UsageUsecase) *UsageHandler {
	return &UsageHandler{usageUsecase: usageUsecase}
}

func (h *UsageHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	overview, err := h.usageUsecase.GetUserOverview(userID)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, overview)
}

func (h *UsageHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 90 {
		days = 7
	}

	stats, err := h.usageUsecase.GetUserStats(userID, days)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, stats)
}

func (h *UsageHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	model := r.URL.Query().Get("model")
	key := r.URL.Query().Get("key")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	logs, total, err := h.usageUsecase.GetUserLogs(userID, page, size, model, key)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Paginated(w, total, page, size, logs)
}

func (h *UsageHandler) GetGlobalOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.usageUsecase.GetGlobalOverview()
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, overview)
}

func (h *UsageHandler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	model := r.URL.Query().Get("model")
	key := r.URL.Query().Get("key")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	logs, total, err := h.usageUsecase.GetAllLogs(page, size, model, key)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Paginated(w, total, page, size, logs)
}

func (h *UsageHandler) GetTopModels(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	models, err := h.usageUsecase.GetTopModels(limit)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, models)
}

func (h *UsageHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	date := time.Now().UTC()
	if value := r.URL.Query().Get("date"); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			resp.Error(w, errcode.ErrInvalidTimeRange)
			return
		}
		date = parsed
	}

	stats, err := h.usageUsecase.GetDailyStats(date)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, stats)
}

func (h *UsageHandler) GetTokenTrend(w http.ResponseWriter, r *http.Request) {
	granularity := r.URL.Query().Get("granularity")
	if granularity != "hour" && granularity != "day" {
		granularity = "day"
	}

	date := time.Now().UTC()
	if value := r.URL.Query().Get("date"); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			resp.Error(w, errcode.ErrInvalidTimeRange)
			return
		}
		date = parsed
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 90 {
		days = 14
	}

	trend, err := h.usageUsecase.GetTokenTrend(granularity, date, days)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}
	resp.Success(w, trend)
}

func (h *UsageHandler) GetTopUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.usageUsecase.GetTopUsers(limit)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}

	resp.Success(w, users)
}

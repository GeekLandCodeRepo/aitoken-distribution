package usecase

import (
	"time"

	"llm-gateway/internal/usage/domain"
)

type UsageUsecase struct {
	logRepo domain.RequestLogRepository
}

func NewUsageUsecase(logRepo domain.RequestLogRepository) *UsageUsecase {
	return &UsageUsecase{logRepo: logRepo}
}

func (uc *UsageUsecase) GetUserOverview(userID string) (*domain.UsageOverview, error) {
	return uc.logRepo.GetUserOverview(userID)
}

func (uc *UsageUsecase) GetUserStats(userID string, days int) (*domain.UsageStatsResponse, error) {
	return uc.logRepo.GetUserStats(userID, days)
}

func (uc *UsageUsecase) GetUserLogs(userID string, page, size int, model string, key string) ([]*domain.RequestLogItem, int64, error) {
	return uc.logRepo.ListByUserID(userID, page, size, model, key)
}

func (uc *UsageUsecase) GetGlobalOverview() (*domain.UsageOverview, error) {
	return uc.logRepo.GetGlobalOverview()
}

func (uc *UsageUsecase) GetAllLogs(page, size int, model string, key string) ([]*domain.RequestLogItem, int64, error) {
	return uc.logRepo.ListAll(page, size, model, key)
}

func (uc *UsageUsecase) GetDailyStats(date time.Time) (*domain.UsageStats, error) {
	return uc.logRepo.GetDailyStats(date)
}

func (uc *UsageUsecase) GetTokenTrend(granularity string, date time.Time, days int) ([]*domain.TokenTrendPoint, error) {
	if granularity != "hour" && granularity != "day" {
		granularity = "day"
	}
	if granularity == "day" && (days < 1 || days > 90) {
		days = 14
	}
	return uc.logRepo.GetTokenTrend(granularity, date, days)
}

func (uc *UsageUsecase) GetTopModels(limit int) ([]*domain.ModelUsageStats, error) {
	return uc.logRepo.GetTopModels(limit)
}

func (uc *UsageUsecase) GetTopUsers(limit int) ([]*domain.UserUsageStats, error) {
	return uc.logRepo.GetTopUsers(limit)
}

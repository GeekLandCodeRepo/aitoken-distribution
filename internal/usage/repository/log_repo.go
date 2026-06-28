package repository

import (
	"fmt"
	"strings"
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/usage/domain"
)

type requestLogRepository struct {
	db *xorm.Engine
}

func NewRequestLogRepository(db *xorm.Engine) domain.RequestLogRepository {
	return &requestLogRepository{db: db}
}

func (r *requestLogRepository) Create(log *domain.RequestLog) error {
	_, err := r.db.Insert(log)
	return err
}

func (r *requestLogRepository) GetByID(id string) (*domain.RequestLog, error) {
	log := &domain.RequestLog{}
	has, err := r.db.ID(id).Get(log)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return log, nil
}

func (r *requestLogRepository) ListByUserID(userID string, page, size int, model string, key string) ([]*domain.RequestLogItem, int64, error) {
	where := "WHERE request_logs.user_id = ?"
	args := []interface{}{userID}
	if model != "" {
		where += " AND request_logs.model = ?"
		args = append(args, model)
	}
	where, args = appendKeyFilter(where, args, key)
	return r.listLogs(where, args, page, size)
}

func (r *requestLogRepository) ListAll(page, size int, model string, key string) ([]*domain.RequestLogItem, int64, error) {
	where := ""
	args := []interface{}{}
	if model != "" {
		where = "WHERE request_logs.model = ?"
		args = append(args, model)
	}
	where, args = appendKeyFilter(where, args, key)
	return r.listLogs(where, args, page, size)
}

func (r *requestLogRepository) listLogs(where string, args []interface{}, page, size int) ([]*domain.RequestLogItem, int64, error) {
	var total int64
	countSQL := `
		SELECT COUNT(*)
		FROM request_logs
		LEFT JOIN api_keys ON api_keys.id = request_logs.api_key_id
		LEFT JOIN users ON users.id = request_logs.user_id
		LEFT JOIN channels ON channels.id = request_logs.channel_id
		` + where
	_, err := r.db.SQL(countSQL, args...).Get(&total)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*domain.RequestLogItem, 0)
	listSQL := `
		SELECT request_logs.*,
		       users.username AS username,
		       users.email AS email,
		       api_keys.name AS key_name,
		       api_keys.key_prefix AS key_prefix,
		       api_keys.key_suffix AS key_suffix,
		       channels.name AS channel
		FROM request_logs
		LEFT JOIN users ON users.id = request_logs.user_id
		LEFT JOIN api_keys ON api_keys.id = request_logs.api_key_id
		LEFT JOIN channels ON channels.id = request_logs.channel_id
		` + where + `
		ORDER BY request_logs.created_at DESC
		LIMIT ? OFFSET ?`
	listArgs := append(append([]interface{}{}, args...), size, (page-1)*size)
	if err := r.db.SQL(listSQL, listArgs...).Find(&logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func appendKeyFilter(where string, args []interface{}, key string) (string, []interface{}) {
	key = strings.TrimSpace(key)
	if key == "" {
		return where, args
	}
	clause := "api_keys.key_prefix LIKE ?"
	args = append(args, key+"%")
	if len(key) >= 6 {
		prefix := key[:6]
		clause = "(api_keys.key_prefix LIKE ?"
		args[len(args)-1] = prefix + "%"
		if len(key) > 6 {
			suffix := key[len(key)-6:]
			clause += " OR api_keys.key_suffix = ?"
			args = append(args, suffix)
		}
		clause += ")"
	}
	if where == "" {
		return "WHERE " + clause, args
	}
	return where + " AND " + clause, args
}

func (r *requestLogRepository) GetUserOverview(userID string) (*domain.UsageOverview, error) {
	overview := &domain.UsageOverview{}

	// 获取用户余额和已用额度
	user := &struct {
		Balance      int64
		UsedQuota    int64
		RequestCount int64
	}{}
	_, err := r.db.Table("users").Select("balance, used_quota, request_count").Where("id = ?", userID).Get(user)
	if err != nil {
		return nil, err
	}
	overview.Balance = user.Balance
	overview.UsedQuota = user.UsedQuota
	overview.RequestCount = user.RequestCount

	// 今日统计
	today := time.Now().UTC().Truncate(24 * time.Hour)
	_, err = r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE user_id = ? AND created_at >= ?
	`, userID, today).Get(&overview.Today.Tokens, &overview.Today.Cost, &overview.Today.Requests)
	if err != nil {
		return nil, err
	}

	// 本月统计
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err = r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE user_id = ? AND created_at >= ?
	`, userID, firstOfMonth).Get(&overview.ThisMonth.Tokens, &overview.ThisMonth.Cost, &overview.ThisMonth.Requests)
	if err != nil {
		return nil, err
	}

	return overview, nil
}

func (r *requestLogRepository) GetUserStats(userID string, days int) (*domain.UsageStatsResponse, error) {
	if days < 1 || days > 90 {
		days = 7
	}
	start := time.Now().UTC().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	stats := make([]*domain.UsageStat, 0)
	if err := r.db.SQL(`
		SELECT to_char(created_at::date, 'YYYY-MM-DD') AS date,
		       COUNT(*) AS requests,
		       COALESCE(SUM(total_tokens), 0) AS tokens,
		       COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		       COALESCE(SUM(cost), 0) AS cost
		FROM request_logs
		WHERE user_id = ? AND created_at >= ?
		GROUP BY created_at::date
		ORDER BY created_at::date ASC
	`, userID, start).Find(&stats); err != nil {
		return nil, err
	}

	byModel := make([]*domain.UsageByModel, 0)
	if err := r.db.SQL(`
		SELECT model,
		       COUNT(*) AS requests,
		       COALESCE(SUM(total_tokens), 0) AS tokens,
		       COALESCE(SUM(cost), 0) AS cost
		FROM request_logs
		WHERE user_id = ? AND created_at >= ?
		GROUP BY model
		ORDER BY tokens DESC
		LIMIT 10
	`, userID, start).Find(&byModel); err != nil {
		return nil, err
	}

	return &domain.UsageStatsResponse{Stats: stats, ByModel: byModel}, nil
}

func (r *requestLogRepository) GetGlobalOverview() (*domain.UsageOverview, error) {
	overview := &domain.UsageOverview{}

	// 全局统计
	total, _ := r.db.Count(&domain.RequestLog{})
	overview.RequestCount = total

	r.db.SQL("SELECT COALESCE(SUM(cost), 0) FROM request_logs").Get(&overview.UsedQuota)

	// 今日统计
	today := time.Now().UTC().Truncate(24 * time.Hour)
	r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE created_at >= ?
	`, today).Get(&overview.Today.Tokens, &overview.Today.Cost, &overview.Today.Requests)

	// 本月统计
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE created_at >= ?
	`, firstOfMonth).Get(&overview.ThisMonth.Tokens, &overview.ThisMonth.Cost, &overview.ThisMonth.Requests)

	return overview, nil
}

func (r *requestLogRepository) GetDailyStats(date time.Time) (*domain.UsageStats, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	stats := &domain.UsageStats{}
	_, err := r.db.SQL(`
		SELECT COUNT(*), COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0)
		FROM request_logs WHERE created_at >= ? AND created_at < ?
	`, start, end).Get(&stats.Requests, &stats.Tokens, &stats.Cost)
	return stats, err
}

func (r *requestLogRepository) GetTokenTrend(granularity string, date time.Time, days int) ([]*domain.TokenTrendPoint, error) {
	if granularity == "hour" {
		return r.getHourlyTokenTrend(date)
	}
	return r.getDailyTokenTrend(days)
}

func (r *requestLogRepository) getHourlyTokenTrend(date time.Time) ([]*domain.TokenTrendPoint, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	dateLabel := start.Format("2006-01-02")

	points := make([]*domain.TokenTrendPoint, 24)
	byHour := make(map[int]*domain.TokenTrendPoint, 24)
	for hour := 0; hour < 24; hour++ {
		point := &domain.TokenTrendPoint{Label: fmt.Sprintf("%02d:00", hour), Date: dateLabel, Hour: hour}
		points[hour] = point
		byHour[hour] = point
	}

	rows := make([]*domain.TokenTrendPoint, 0)
	if err := r.db.SQL(`
		SELECT EXTRACT(HOUR FROM created_at)::int AS hour,
		       COUNT(*) AS requests,
		       COALESCE(SUM(total_tokens), 0) AS tokens,
		       COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		       COALESCE(SUM(reasoning_tokens), 0) AS reasoning_tokens,
		       COALESCE(SUM(cache_tokens), 0) AS cache_tokens,
		       COALESCE(SUM(cost), 0) AS cost
		FROM request_logs
		WHERE created_at >= ? AND created_at < ?
		GROUP BY EXTRACT(HOUR FROM created_at)::int
		ORDER BY hour ASC
	`, start, end).Find(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if point := byHour[row.Hour]; point != nil {
			point.Requests = row.Requests
			point.Tokens = row.Tokens
			point.PromptTokens = row.PromptTokens
			point.CompletionTokens = row.CompletionTokens
			point.ReasoningTokens = row.ReasoningTokens
			point.CacheTokens = row.CacheTokens
			point.Cost = row.Cost
		}
	}
	return points, nil
}

func (r *requestLogRepository) getDailyTokenTrend(days int) ([]*domain.TokenTrendPoint, error) {
	if days < 1 || days > 90 {
		days = 14
	}
	start := time.Now().UTC().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, days)

	points := make([]*domain.TokenTrendPoint, 0, days)
	byDate := make(map[string]*domain.TokenTrendPoint, days)
	for i := 0; i < days; i++ {
		dateLabel := start.AddDate(0, 0, i).Format("2006-01-02")
		point := &domain.TokenTrendPoint{Label: dateLabel, Date: dateLabel, Hour: -1}
		points = append(points, point)
		byDate[dateLabel] = point
	}

	rows := make([]*domain.TokenTrendPoint, 0)
	if err := r.db.SQL(`
		SELECT to_char(created_at::date, 'YYYY-MM-DD') AS date,
		       COUNT(*) AS requests,
		       COALESCE(SUM(total_tokens), 0) AS tokens,
		       COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		       COALESCE(SUM(reasoning_tokens), 0) AS reasoning_tokens,
		       COALESCE(SUM(cache_tokens), 0) AS cache_tokens,
		       COALESCE(SUM(cost), 0) AS cost
		FROM request_logs
		WHERE created_at >= ? AND created_at < ?
		GROUP BY created_at::date
		ORDER BY created_at::date ASC
	`, start, end).Find(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if point := byDate[row.Date]; point != nil {
			point.Requests = row.Requests
			point.Tokens = row.Tokens
			point.PromptTokens = row.PromptTokens
			point.CompletionTokens = row.CompletionTokens
			point.ReasoningTokens = row.ReasoningTokens
			point.CacheTokens = row.CacheTokens
			point.Cost = row.Cost
		}
	}
	return points, nil
}

func (r *requestLogRepository) GetTopModels(limit int) ([]*domain.ModelUsageStats, error) {
	if limit <= 0 {
		limit = 10
	}

	var totalTokens int64
	if _, err := r.db.SQL("SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs").Get(&totalTokens); err != nil {
		return nil, err
	}

	stats := make([]*domain.ModelUsageStats, 0)
	if err := r.db.SQL(`
		SELECT model,
		       COUNT(*) AS requests,
		       COALESCE(SUM(total_tokens), 0) AS tokens,
		       COALESCE(SUM(cost), 0) AS cost
		FROM request_logs
		GROUP BY model
		ORDER BY tokens DESC
		LIMIT ?
	`, limit).Find(&stats); err != nil {
		return nil, err
	}

	if totalTokens > 0 {
		for _, item := range stats {
			item.Percentage = float64(item.Tokens) / float64(totalTokens) * 100
		}
	}

	return stats, nil
}

func (r *requestLogRepository) GetTopUsers(limit int) ([]*domain.UserUsageStats, error) {
	if limit <= 0 {
		limit = 10
	}

	var totalTokens int64
	if _, err := r.db.SQL("SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs").Get(&totalTokens); err != nil {
		return nil, err
	}

	stats := make([]*domain.UserUsageStats, 0)
	if err := r.db.SQL(`
		SELECT rl.user_id,
		       u.username,
		       u.email,
		       COUNT(*) AS requests,
		       COALESCE(SUM(rl.total_tokens), 0) AS tokens,
		       COALESCE(SUM(rl.cost), 0) AS cost
		FROM request_logs rl
		LEFT JOIN users u ON u.id = rl.user_id
		GROUP BY rl.user_id, u.username, u.email
		ORDER BY tokens DESC
		LIMIT ?
	`, limit).Find(&stats); err != nil {
		return nil, err
	}

	if totalTokens > 0 {
		for _, item := range stats {
			item.Percentage = float64(item.Tokens) / float64(totalTokens) * 100
		}
	}

	return stats, nil
}

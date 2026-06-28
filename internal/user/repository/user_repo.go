package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/shared/cache"
	"llm-gateway/internal/shared/uuid"
	"llm-gateway/internal/user/domain"
)

type userRepository struct {
	db *xorm.Engine
}

const defaultUserChannelIDsKey = "default_user_channel_ids"

func NewUserRepository(db *xorm.Engine) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	_, err := r.db.Insert(user)
	return err
}

func (r *userRepository) GetByID(id string) (*domain.User, error) {
	user := &domain.User{}
	has, err := r.db.ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	has, err := r.db.Where("email = ?", email).Get(user)
	if err != nil {
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	user := &domain.User{}
	has, err := r.db.Where("username = ?", username).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	_, err := r.db.ID(user.ID).Update(user)
	return err
}

func (r *userRepository) Delete(id string) error {
	_, err := r.db.ID(id).Delete(&domain.User{})
	return err
}

func (r *userRepository) List(page, size int, search string, status, role *int) ([]*domain.User, int64, error) {
	var users []*domain.User
	query := r.db.Limit(size, (page-1)*size)

	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if role != nil {
		query = query.Where("role = ?", *role)
	}

	total, err := query.Desc("created_at").FindAndCount(&users)
	if err != nil {
		return nil, 0, err
	}
	if len(users) > 0 {
		if err := r.fillUsageFromLogs(users); err != nil {
			return nil, 0, err
		}
	}

	return users, total, nil
}

func (r *userRepository) fillUsageFromLogs(users []*domain.User) error {
	ids := make([]string, 0, len(users))
	byID := make(map[string]*domain.User, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
		byID[user.ID] = user
	}

	type usageRow struct {
		UserID       string `xorm:"user_id"`
		UsedQuota    int64  `xorm:"used_quota"`
		RequestCount int64  `xorm:"request_count"`
	}

	rows := make([]usageRow, 0, len(users))
	err := r.db.Table("request_logs").
		Select("user_id, COALESCE(SUM(cost), 0) AS used_quota, COUNT(*) AS request_count").
		In("user_id", ids).
		GroupBy("user_id").
		Find(&rows)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if user := byID[row.UserID]; user != nil {
			user.UsedQuota = row.UsedQuota
			user.RequestCount = row.RequestCount
		}
	}
	return nil
}

func (r *userRepository) UpdateBalance(id string, amount int64) error {
	user := &domain.User{Balance: amount}
	_, err := r.db.ID(id).Cols("balance").Update(user)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) IncrementBalance(id string, amount int64) error {
	_, err := r.db.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, id)
	return err
}

func (r *userRepository) UpdateUsedQuota(id string, amount int64) error {
	user := &domain.User{UsedQuota: amount, UpdatedAt: time.Now()}
	_, err := r.db.ID(id).Cols("used_quota", "updated_at").Update(user)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) IncrementRequestCount(id string) error {
	_, err := r.db.Exec("UPDATE users SET request_count = request_count + 1 WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) CreateInviteCode(code *domain.InviteCode) error {
	data := map[string]interface{}{
		"id":             code.ID,
		"code":           code.Code,
		"reward_amount":  code.RewardAmount,
		"new_user_bonus": code.NewUserBonus,
		"created_at":     code.CreatedAt,
	}
	if code.InviterUserID != "" {
		data["inviter_user_id"] = code.InviterUserID
	}
	_, err := r.db.Table("invite_codes").Insert(data)
	return err
}

func (r *userRepository) ListInviteCodes(page, size int) ([]*domain.InviteCode, int64, error) {
	var codes []*domain.InviteCode
	total, err := r.db.Limit(size, (page-1)*size).Desc("created_at").FindAndCount(&codes)
	return codes, total, err
}

func (r *userRepository) UseInviteCode(code string, userID string) (*domain.InviteCode, error) {
	session := r.db.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	rows, err := session.QueryString(`
		SELECT id, COALESCE(inviter_user_id::text, '') AS inviter_user_id, reward_amount, new_user_bonus
		FROM invite_codes
		WHERE code = ? AND used_by_user_id IS NULL AND used_at IS NULL
		FOR UPDATE
	`, code)
	if err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if len(rows) == 0 {
		_ = session.Rollback()
		return nil, fmt.Errorf("invite code unavailable")
	}
	rewardAmount := int64(0)
	_, _ = fmt.Sscan(rows[0]["reward_amount"], &rewardAmount)
	newUserBonus := int64(0)
	_, _ = fmt.Sscan(rows[0]["new_user_bonus"], &newUserBonus)
	invite := &domain.InviteCode{
		ID:            rows[0]["id"],
		Code:          code,
		InviterUserID: rows[0]["inviter_user_id"],
		RewardAmount:  rewardAmount,
		NewUserBonus:  newUserBonus,
	}

	now := time.Now()
	invite.UsedByUserID = userID
	invite.UsedAt = &now
	if _, err := session.Exec("UPDATE invite_codes SET used_by_user_id = ?, used_at = ? WHERE id = ?", userID, now, invite.ID); err != nil {
		_ = session.Rollback()
		return nil, err
	}

	if invite.InviterUserID != "" && invite.RewardAmount > 0 {
		if _, err := session.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", invite.RewardAmount, invite.InviterUserID); err != nil {
			_ = session.Rollback()
			return nil, err
		}
	}
	if invite.NewUserBonus > 0 {
		if _, err := session.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", invite.NewUserBonus, userID); err != nil {
			_ = session.Rollback()
			return nil, err
		}
	}

	return invite, session.Commit()
}

func (r *userRepository) GetSetting(key string) (string, error) {
	setting := &domain.AppSetting{}
	has, err := r.db.ID(key).Get(setting)
	if err != nil || !has {
		return "", err
	}
	return setting.Value, nil
}

func (r *userRepository) SetSetting(key string, value string) error {
	setting := &domain.AppSetting{Key: key, Value: value, UpdatedAt: time.Now()}
	affected, err := r.db.ID(key).Cols("value", "updated_at").Update(setting)
	if err != nil {
		return err
	}
	if affected == 0 {
		_, err = r.db.Insert(setting)
	}
	return err
}

func (r *userRepository) ListUserChannelOptions(userID string) ([]*domain.UserChannelOption, error) {
	type channelOptionRow struct {
		ID      string `xorm:"'id'"`
		Name    string `xorm:"'name'"`
		Type    int    `xorm:"'type'"`
		Status  int    `xorm:"'status'"`
		BaseURL string `xorm:"'base_url'"`
		Models  string `xorm:"'models'"`
	}

	allowedIDs, err := r.ListAllowedChannelIDs(userID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(allowedIDs))
	for _, channelID := range allowedIDs {
		allowed[channelID] = struct{}{}
	}

	rows := make([]*channelOptionRow, 0)
	err = r.db.Table("channels").Alias("c").
		Select("c.id, c.name, c.type, c.status, c.base_url, c.models").
		Desc("c.priority").Desc("c.weight").Desc("c.created_at").
		Find(&rows)
	if err != nil {
		return nil, err
	}

	options := make([]*domain.UserChannelOption, 0, len(rows))
	for _, row := range rows {
		models := []string{}
		if row.Models != "" {
			_ = json.Unmarshal([]byte(row.Models), &models)
		}
		options = append(options, &domain.UserChannelOption{
			ID:      row.ID,
			Name:    row.Name,
			Type:    row.Type,
			Status:  row.Status,
			BaseURL: row.BaseURL,
			Models:  models,
			Allowed: hasString(allowed, row.ID),
		})
	}
	return options, nil
}

func (r *userRepository) ReplaceUserChannels(userID string, channelIDs []string) error {
	session := r.db.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}
	if _, err := session.Where("user_id = ?", userID).Delete(&domain.UserChannelPermission{}); err != nil {
		_ = session.Rollback()
		return err
	}

	seen := make(map[string]struct{}, len(channelIDs))
	permissions := make([]*domain.UserChannelPermission, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID == "" {
			continue
		}
		if _, ok := seen[channelID]; ok {
			continue
		}
		seen[channelID] = struct{}{}
		permissions = append(permissions, &domain.UserChannelPermission{
			ID:        uuid.NewV7String(),
			UserID:    userID,
			ChannelID: channelID,
			CreatedAt: time.Now(),
		})
	}
	if len(permissions) > 0 {
		if _, err := session.Insert(permissions); err != nil {
			_ = session.Rollback()
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}
	cache.DeletePattern(context.Background(), fmt.Sprintf("channel_candidates:%s:*", userID))
	return nil
}

func replaceUserChannelsWithSession(session *xorm.Session, userID string, channelIDs []string) error {
	if _, err := session.Where("user_id = ?", userID).Delete(&domain.UserChannelPermission{}); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(channelIDs))
	permissions := make([]*domain.UserChannelPermission, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID == "" {
			continue
		}
		if _, ok := seen[channelID]; ok {
			continue
		}
		seen[channelID] = struct{}{}
		permissions = append(permissions, &domain.UserChannelPermission{
			ID:        uuid.NewV7String(),
			UserID:    userID,
			ChannelID: channelID,
			CreatedAt: time.Now(),
		})
	}
	if len(permissions) > 0 {
		_, err := session.Insert(permissions)
		return err
	}
	return nil
}

func (r *userRepository) ListAllowedChannelIDs(userID string) ([]string, error) {
	permissions := make([]*domain.UserChannelPermission, 0)
	if err := r.db.Where("user_id = ?", userID).Find(&permissions); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		ids = append(ids, permission.ChannelID)
	}
	return ids, nil
}

func (r *userRepository) GetDefaultUserChannelIDs() ([]string, error) {
	value, err := r.GetSetting(defaultUserChannelIDsKey)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return []string{}, nil
	}
	ids := []string{}
	if err := json.Unmarshal([]byte(value), &ids); err != nil {
		return nil, err
	}
	return dedupeChannelIDs(ids), nil
}

func (r *userRepository) SaveDefaultUserChannelIDs(channelIDs []string) error {
	data, err := json.Marshal(dedupeChannelIDs(channelIDs))
	if err != nil {
		return err
	}
	return r.SetSetting(defaultUserChannelIDsKey, string(data))
}

func (r *userRepository) ListDefaultUserChannelOptions() ([]*domain.UserChannelOption, error) {
	channelIDs, err := r.GetDefaultUserChannelIDs()
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(channelIDs))
	for _, channelID := range channelIDs {
		allowed[channelID] = struct{}{}
	}

	type channelOptionRow struct {
		ID      string `xorm:"'id'"`
		Name    string `xorm:"'name'"`
		Type    int    `xorm:"'type'"`
		Status  int    `xorm:"'status'"`
		BaseURL string `xorm:"'base_url'"`
		Models  string `xorm:"'models'"`
	}
	rows := make([]*channelOptionRow, 0)
	if err := r.db.Table("channels").Alias("c").
		Select("c.id, c.name, c.type, c.status, c.base_url, c.models").
		Desc("c.priority").Desc("c.weight").Desc("c.created_at").
		Find(&rows); err != nil {
		return nil, err
	}

	options := make([]*domain.UserChannelOption, 0, len(rows))
	for _, row := range rows {
		models := []string{}
		if row.Models != "" {
			_ = json.Unmarshal([]byte(row.Models), &models)
		}
		_, isAllowed := allowed[row.ID]
		options = append(options, &domain.UserChannelOption{
			ID:      row.ID,
			Name:    row.Name,
			Type:    row.Type,
			Status:  row.Status,
			BaseURL: row.BaseURL,
			Models:  models,
			Allowed: isAllowed,
		})
	}
	return options, nil
}

func (r *userRepository) ApplyDefaultChannelsToUser(userID string) error {
	channelIDs, err := r.GetDefaultUserChannelIDs()
	if err != nil {
		return err
	}
	return r.ReplaceUserChannels(userID, channelIDs)
}

func (r *userRepository) ApplyDefaultChannelsToAllUsers() (int64, error) {
	channelIDs, err := r.GetDefaultUserChannelIDs()
	if err != nil {
		return 0, err
	}

	session := r.db.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return 0, err
	}

	var users []*domain.User
	if err := session.Find(&users); err != nil {
		_ = session.Rollback()
		return 0, err
	}
	for _, user := range users {
		if err := replaceUserChannelsWithSession(session, user.ID, channelIDs); err != nil {
			_ = session.Rollback()
			return 0, err
		}
	}
	if err := session.Commit(); err != nil {
		return 0, err
	}
	cache.DeletePattern(context.Background(), "channel_candidates:*")
	return int64(len(users)), nil
}

func dedupeChannelIDs(channelIDs []string) []string {
	seen := make(map[string]struct{}, len(channelIDs))
	result := make([]string, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID == "" {
			continue
		}
		if _, ok := seen[channelID]; ok {
			continue
		}
		seen[channelID] = struct{}{}
		result = append(result, channelID)
	}
	return result
}

func hasString(values map[string]struct{}, value string) bool {
	_, ok := values[value]
	return ok
}

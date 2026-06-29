package domain

import (
	"errors"
	"time"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type User struct {
	// ID is the user's UUID primary key.
	ID string `json:"id" xorm:"'id' uuid pk notnull comment('用户ID')"`
	// Username is the unique display/login name.
	Username string `json:"username" xorm:"varchar(64) unique notnull comment('用户名')"`
	// Email is the user's unique email address.
	Email string `json:"email" xorm:"varchar(255) unique notnull comment('邮箱')"`
	// PasswordHash stores the Argon2id password hash.
	PasswordHash string `json:"-" xorm:"varchar(255) notnull comment('Argon2id密码哈希')"`
	// Role stores the user role, where 1 is normal user and 10 is admin.
	Role int `json:"role" xorm:"smallint default(1) comment('角色')"`
	// Balance stores remaining quota in internal units.
	Balance int64 `json:"balance" xorm:"bigint default(0) comment('余额')"`
	// UsedQuota stores consumed quota in internal units.
	UsedQuota int64 `json:"used_quota" xorm:"bigint default(0) comment('已用额度')"`
	// RequestCount stores total request count.
	RequestCount int64 `json:"request_count" xorm:"bigint default(0) comment('请求次数')"`
	// Status stores account status, where 1 is enabled and 0 is disabled.
	Status int `json:"status" xorm:"smallint default(1) comment('状态')"`
	// GroupName stores the user group for routing and permissions.
	GroupName string `json:"group_name" xorm:"varchar(32) default('default') comment('用户组')"`
	// CreatedAt stores the creation time.
	CreatedAt time.Time `json:"created_at" xorm:"created comment('创建时间')"`
	// UpdatedAt stores the last update time.
	UpdatedAt time.Time `json:"updated_at" xorm:"updated comment('更新时间')"`
}

type InviteCode struct {
	// ID is the invite code UUID primary key.
	ID string `json:"id" xorm:"'id' uuid pk notnull comment('邀请码ID')"`
	// Code is the one-time invite code shown to users.
	Code string `json:"code" xorm:"varchar(64) unique notnull comment('邀请码')"`
	// InviterUserID is the user who owns this invite code; empty means admin-generated without owner.
	InviterUserID string `json:"inviter_user_id" xorm:"uuid inviter_user_id comment('邀请人用户ID')"`
	// UsedByUserID is the user who consumed this invite code.
	UsedByUserID string `json:"used_by_user_id" xorm:"uuid used_by_user_id comment('使用人用户ID')"`
	// UsedAt stores when the invite code was consumed.
	UsedAt *time.Time `json:"used_at" xorm:"datetime comment('使用时间')"`
	// RewardAmount is the inviter cashback amount in internal units.
	RewardAmount int64 `json:"reward_amount" xorm:"bigint default(0) comment('邀请人返现金额')"`
	// NewUserBonus is the bonus granted to the newly registered user in internal units.
	NewUserBonus int64 `json:"new_user_bonus" xorm:"bigint default(0) comment('新用户赠送金额')"`
	// CreatedAt stores the invite code creation time.
	CreatedAt time.Time `json:"created_at" xorm:"created comment('创建时间')"`
}

func (InviteCode) TableName() string {
	return "invite_codes"
}

type AppSetting struct {
	// Key is the unique setting key.
	Key string `json:"key" xorm:"varchar(128) pk notnull comment('设置键')"`
	// Value stores the setting value as text.
	Value string `json:"value" xorm:"text notnull comment('设置值')"`
	// UpdatedAt stores the last setting update time.
	UpdatedAt time.Time `json:"updated_at" xorm:"updated comment('更新时间')"`
}

type UserChannelPermission struct {
	// ID is the permission UUID primary key.
	ID string `json:"id" xorm:"'id' uuid pk notnull comment('用户渠道权限ID')"`
	// UserID is the user that owns this channel permission.
	UserID string `json:"user_id" xorm:"'user_id' uuid notnull unique(ux_user_channel_permission) index comment('用户ID')"`
	// ChannelID is the allowed channel for the user.
	ChannelID string `json:"channel_id" xorm:"'channel_id' uuid notnull unique(ux_user_channel_permission) index comment('渠道ID')"`
	// CreatedAt stores the creation time.
	CreatedAt time.Time `json:"created_at" xorm:"created comment('创建时间')"`
}

func (UserChannelPermission) TableName() string {
	return "user_channel_permissions"
}

type UserChannelOption struct {
	ID      string   `json:"id" xorm:"'id'"`
	Name    string   `json:"name" xorm:"'name'"`
	Type    int      `json:"type" xorm:"'type'"`
	Status  int      `json:"status" xorm:"'status'"`
	BaseURL string   `json:"base_url" xorm:"'base_url'"`
	Models  []string `json:"models"`
	Allowed bool     `json:"allowed" xorm:"'allowed'"`
}

func (AppSetting) TableName() string {
	return "app_settings"
}

func (User) TableName() string {
	return "users"
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(page, size int, search string, status, role *int) ([]*User, int64, error)
	UpdateBalance(id string, amount int64) error
	IncrementBalance(id string, amount int64) error
	AdminAdjustBalance(id string, amount int64, txType int, description string) (int64, int64, error)
	UpdateUsedQuota(id string, amount int64) error
	IncrementRequestCount(id string) error
	CreateInviteCode(code *InviteCode) error
	ListInviteCodes(page, size int) ([]*InviteCode, int64, error)
	UseInviteCode(code string, userID string) (*InviteCode, error)
	GetSetting(key string) (string, error)
	SetSetting(key string, value string) error
	ListUserChannelOptions(userID string) ([]*UserChannelOption, error)
	ReplaceUserChannels(userID string, channelIDs []string) error
	ListAllowedChannelIDs(userID string) ([]string, error)
	GetDefaultUserChannelIDs() ([]string, error)
	SaveDefaultUserChannelIDs(channelIDs []string) error
	ListDefaultUserChannelOptions() ([]*UserChannelOption, error)
	ApplyDefaultChannelsToUser(userID string) error
	ApplyDefaultChannelsToAllUsers() (int64, error)
}

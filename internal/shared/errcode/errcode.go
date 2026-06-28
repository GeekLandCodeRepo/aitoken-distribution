package errcode

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	// System 10xxxx
	ErrInvalidBody        = &AppError{Code: 100101, Message: "request body invalid", HTTP: 400}
	ErrBodyTooLarge       = &AppError{Code: 100102, Message: "request body too large", HTTP: 400}
	ErrContentType        = &AppError{Code: 100103, Message: "unsupported content type", HTTP: 400}
	ErrUnauthorized       = &AppError{Code: 100201, Message: "unauthorized", HTTP: 401}
	ErrTokenExpired       = &AppError{Code: 100202, Message: "token expired", HTTP: 401}
	ErrTokenInvalid       = &AppError{Code: 100203, Message: "token invalid", HTTP: 401}
	ErrForbidden          = &AppError{Code: 100204, Message: "forbidden", HTTP: 403}
	ErrNotFound           = &AppError{Code: 100301, Message: "not found", HTTP: 404}
	ErrMethodNotAllowed   = &AppError{Code: 100401, Message: "method not allowed", HTTP: 405}
	ErrRateLimit          = &AppError{Code: 100402, Message: "rate limit exceeded", HTTP: 429}
	ErrInternal           = &AppError{Code: 100501, Message: "internal error", HTTP: 500}
	ErrDatabase           = &AppError{Code: 100502, Message: "database error", HTTP: 500}
	ErrRedis              = &AppError{Code: 100503, Message: "redis error", HTTP: 500}
	ErrServiceUnavailable = &AppError{Code: 100504, Message: "service unavailable", HTTP: 503}

	// Auth 20xxxx
	ErrInvalidEmail    = &AppError{Code: 200101, Message: "invalid email format", HTTP: 400}
	ErrInvalidUsername = &AppError{Code: 200102, Message: "invalid username format", HTTP: 400}
	ErrInvalidPassword = &AppError{Code: 200103, Message: "invalid password format", HTTP: 400}
	ErrInvalidCode     = &AppError{Code: 200104, Message: "invalid verification code", HTTP: 400}
	ErrEmailExists     = &AppError{Code: 200301, Message: "email already exists", HTTP: 409}
	ErrUsernameExists  = &AppError{Code: 200302, Message: "username already exists", HTTP: 409}
	ErrBadCredentials  = &AppError{Code: 200303, Message: "invalid email or password", HTTP: 401}
	ErrUserDisabled    = &AppError{Code: 200304, Message: "user account disabled", HTTP: 403}
	ErrRefreshInvalid  = &AppError{Code: 200305, Message: "refresh token invalid", HTTP: 401}
	ErrWrongPassword   = &AppError{Code: 200401, Message: "current password incorrect", HTTP: 400}
	ErrSamePassword    = &AppError{Code: 200402, Message: "new password same as current", HTTP: 400}
	ErrInviteRequired  = &AppError{Code: 200403, Message: "invite code required", HTTP: 400}
	ErrInviteInvalid   = &AppError{Code: 200404, Message: "invite code invalid or used", HTTP: 400}

	// User 30xxxx
	ErrInvalidUserID       = &AppError{Code: 300101, Message: "invalid user id", HTTP: 400}
	ErrInvalidPagination   = &AppError{Code: 300102, Message: "invalid pagination params", HTTP: 400}
	ErrInvalidSort         = &AppError{Code: 300103, Message: "invalid sort params", HTTP: 400}
	ErrUserNotFound        = &AppError{Code: 300301, Message: "user not found", HTTP: 404}
	ErrUserAlreadyDisabled = &AppError{Code: 300302, Message: "user already disabled", HTTP: 409}
	ErrUserDeleted         = &AppError{Code: 300303, Message: "user deleted", HTTP: 409}
	ErrInvalidAmount       = &AppError{Code: 300401, Message: "invalid amount", HTTP: 400}
	ErrInsufficientBalance = &AppError{Code: 300402, Message: "insufficient balance", HTTP: 400}
	ErrQuotaExceeded       = &AppError{Code: 300403, Message: "quota exceeded", HTTP: 400}

	// API Key 40xxxx
	ErrKeyNameRequired   = &AppError{Code: 400101, Message: "key name is required", HTTP: 400}
	ErrKeyNameTooLong    = &AppError{Code: 400102, Message: "key name too long", HTTP: 400}
	ErrInvalidQuotaLimit = &AppError{Code: 400103, Message: "invalid quota limit", HTTP: 400}
	ErrInvalidRateLimit  = &AppError{Code: 400104, Message: "invalid rate limit", HTTP: 400}
	ErrInvalidModels     = &AppError{Code: 400105, Message: "invalid models format", HTTP: 400}
	ErrInvalidIPs        = &AppError{Code: 400106, Message: "invalid ip whitelist format", HTTP: 400}
	ErrInvalidExpiry     = &AppError{Code: 400107, Message: "invalid expiry time", HTTP: 400}
	ErrKeyNotFound       = &AppError{Code: 400301, Message: "api key not found", HTTP: 404}
	ErrKeyDisabled       = &AppError{Code: 400302, Message: "api key disabled", HTTP: 409}
	ErrKeyExpired        = &AppError{Code: 400303, Message: "api key expired", HTTP: 409}
	ErrKeyQuotaUsed      = &AppError{Code: 400304, Message: "api key quota exhausted", HTTP: 409}
	ErrModelNotAllowed   = &AppError{Code: 400401, Message: "model not allowed for this key", HTTP: 403}
	ErrIPNotAllowed      = &AppError{Code: 400402, Message: "ip address not allowed", HTTP: 403}
	ErrKeyRateLimited    = &AppError{Code: 400403, Message: "key rate limit exceeded", HTTP: 429}
	ErrKeyNotOwned       = &AppError{Code: 400404, Message: "key not owned by user", HTTP: 400}

	// Channel 50xxxx
	ErrChanNameRequired = &AppError{Code: 500101, Message: "channel name required", HTTP: 400}
	ErrChanTypeInvalid  = &AppError{Code: 500102, Message: "invalid channel type", HTTP: 400}
	ErrChanURLInvalid   = &AppError{Code: 500103, Message: "invalid base url", HTTP: 400}
	ErrChanKeyEmpty     = &AppError{Code: 500104, Message: "api key is empty", HTTP: 400}
	ErrChanModelsEmpty  = &AppError{Code: 500105, Message: "models list is empty", HTTP: 400}
	ErrChanPriority     = &AppError{Code: 500106, Message: "invalid priority value", HTTP: 400}
	ErrChanWeight       = &AppError{Code: 500107, Message: "invalid weight value", HTTP: 400}
	ErrChanGroups       = &AppError{Code: 500108, Message: "invalid groups format", HTTP: 400}
	ErrChanNotFound     = &AppError{Code: 500301, Message: "channel not found", HTTP: 404}
	ErrChanDisabled     = &AppError{Code: 500302, Message: "channel disabled", HTTP: 409}
	ErrChanTestFailed   = &AppError{Code: 500401, Message: "channel test failed", HTTP: 502}
	ErrChanTimeout      = &AppError{Code: 500402, Message: "channel connection timeout", HTTP: 502}
	ErrChanAuthFailed   = &AppError{Code: 500403, Message: "channel authentication failed", HTTP: 502}
	ErrChanInUse        = &AppError{Code: 500404, Message: "channel in use, cannot delete", HTTP: 409}
	ErrBatchLimit       = &AppError{Code: 500405, Message: "batch limit exceeded (max 100)", HTTP: 400}

	// Pricing 60xxxx
	ErrModelRequired            = &AppError{Code: 600101, Message: "model name required", HTTP: 400}
	ErrInvalidChanType          = &AppError{Code: 600102, Message: "invalid channel type", HTTP: 400}
	ErrInvalidPromptPrice       = &AppError{Code: 600103, Message: "invalid prompt price", HTTP: 400}
	ErrInvalidCompPrice         = &AppError{Code: 600104, Message: "invalid completion price", HTTP: 400}
	ErrInvalidCachedPromptPrice = &AppError{Code: 600105, Message: "invalid cached prompt price", HTTP: 400}
	ErrPricingNotFound          = &AppError{Code: 600301, Message: "pricing not found", HTTP: 404}
	ErrPricingExists            = &AppError{Code: 600302, Message: "pricing already exists", HTTP: 409}
	ErrSyncFailed               = &AppError{Code: 600401, Message: "sync failed, no channels available", HTTP: 400}

	// Billing 70xxxx
	ErrInvalidRedeemCode   = &AppError{Code: 700101, Message: "invalid redeem code format", HTTP: 400}
	ErrInvalidTopup        = &AppError{Code: 700102, Message: "invalid topup amount", HTTP: 400}
	ErrInvalidBatchCount   = &AppError{Code: 700103, Message: "invalid batch count (1-1000)", HTTP: 400}
	ErrInvalidExpiryTime   = &AppError{Code: 700104, Message: "invalid expiry time", HTTP: 400}
	ErrCodeNotFound        = &AppError{Code: 700301, Message: "redeem code not found", HTTP: 404}
	ErrCodeUsed            = &AppError{Code: 700302, Message: "redeem code already used", HTTP: 409}
	ErrCodeExpired         = &AppError{Code: 700303, Message: "redeem code expired", HTTP: 409}
	ErrBalanceInsufficient = &AppError{Code: 700401, Message: "insufficient balance", HTTP: 402}
	ErrDuplicateCharge     = &AppError{Code: 700402, Message: "duplicate charge detected", HTTP: 409}
	ErrChargeFailed        = &AppError{Code: 700403, Message: "charge failed", HTTP: 500}
	ErrInvalidRefund       = &AppError{Code: 700404, Message: "invalid refund amount", HTTP: 400}
	ErrTxnNotFound         = &AppError{Code: 700405, Message: "transaction not found", HTTP: 409}

	// Relay 80xxxx
	ErrInvalidRelayBody  = &AppError{Code: 800101, Message: "invalid request body", HTTP: 400}
	ErrModelMissing      = &AppError{Code: 800102, Message: "model parameter required", HTTP: 400}
	ErrInvalidMessages   = &AppError{Code: 800103, Message: "invalid messages format", HTTP: 400}
	ErrUnsupportedParam  = &AppError{Code: 800104, Message: "unsupported parameter", HTTP: 400}
	ErrModelNotAvail     = &AppError{Code: 800301, Message: "model not available", HTTP: 404}
	ErrNoChannel         = &AppError{Code: 800302, Message: "no available channel", HTTP: 404}
	ErrAllChanFailed     = &AppError{Code: 800303, Message: "all channels failed", HTTP: 502}
	ErrUpstreamTimeout   = &AppError{Code: 800304, Message: "upstream timeout", HTTP: 504}
	ErrUpstreamInvalid   = &AppError{Code: 800305, Message: "upstream invalid response", HTTP: 502}
	ErrUpstreamAuth      = &AppError{Code: 800306, Message: "upstream auth failed", HTTP: 502}
	ErrUpstreamRateLimit = &AppError{Code: 800307, Message: "upstream rate limited", HTTP: 429}
	ErrStreamBroken      = &AppError{Code: 800401, Message: "stream interrupted", HTTP: 500}
	ErrTokenCountFailed  = &AppError{Code: 800402, Message: "token count failed", HTTP: 500}
	ErrFormatConvert     = &AppError{Code: 800403, Message: "format conversion failed", HTTP: 500}

	// Usage 90xxxx
	ErrInvalidTimeRange  = &AppError{Code: 900101, Message: "invalid time range", HTTP: 400}
	ErrTimeRangeTooLarge = &AppError{Code: 900102, Message: "time range too large (max 90 days)", HTTP: 400}
	ErrInvalidGroupBy    = &AppError{Code: 900103, Message: "invalid group_by parameter", HTTP: 400}
	ErrInvalidExportFmt  = &AppError{Code: 900104, Message: "invalid export format", HTTP: 400}
	ErrLogNotFound       = &AppError{Code: 900301, Message: "log record not found", HTTP: 404}
)

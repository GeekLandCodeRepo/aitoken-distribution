package resp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"llm-gateway/internal/shared/errcode"
)

type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type openAIErrorResponse struct {
	Error openAIError `json:"error"`
}

type openAIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    string  `json:"code"`
}

func Success(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Code: 0, Data: data})
}

func Created(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Code: 0, Data: data})
}

func Error(w http.ResponseWriter, err *errcode.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTP)
	json.NewEncoder(w).Encode(Response{Code: err.Code, Message: err.Message})
}

func OpenAIError(w http.ResponseWriter, err *errcode.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTP)
	json.NewEncoder(w).Encode(openAIErrorResponse{
		Error: openAIError{
			Message: err.Message,
			Type:    openAIErrorType(err),
			Param:   nil,
			Code:    fmt.Sprintf("%d", err.Code),
		},
	})
}

func ErrorWithDetail(w http.ResponseWriter, err *errcode.AppError, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTP)
	json.NewEncoder(w).Encode(Response{Code: err.Code, Message: err.Message, Data: detail})
}

func openAIErrorType(err *errcode.AppError) string {
	switch err {
	case errcode.ErrUnauthorized, errcode.ErrTokenExpired, errcode.ErrTokenInvalid, errcode.ErrRefreshInvalid:
		return "authentication_error"
	case errcode.ErrForbidden, errcode.ErrModelNotAllowed, errcode.ErrIPNotAllowed:
		return "permission_error"
	case errcode.ErrRateLimit, errcode.ErrKeyRateLimited:
		return "rate_limit_error"
	case errcode.ErrBalanceInsufficient, errcode.ErrKeyQuotaUsed, errcode.ErrQuotaExceeded:
		return "insufficient_quota"
	case errcode.ErrInternal, errcode.ErrDatabase, errcode.ErrRedis, errcode.ErrServiceUnavailable:
		return "server_error"
	default:
		return "invalid_request_error"
	}
}

func Paginated(w http.ResponseWriter, total int64, page, size int, items any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Code: 0,
		Data: map[string]any{
			"total": total,
			"page":  page,
			"size":  size,
			"items": items,
		},
	})
}

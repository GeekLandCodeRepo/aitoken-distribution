package handler

import (
	"encoding/json"
	"net/http"

	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/resp"
	"llm-gateway/internal/user/usecase"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req usecase.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	user, err := h.authUsecase.Register(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecase.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	tokenResp, err := h.authUsecase.Login(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, tokenResp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	tokenResp, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, tokenResp)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.authUsecase.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]interface{}{
		"id":            user.ID,
		"email":         user.Email,
		"username":      user.Username,
		"role":          user.Role,
		"balance":       user.Balance,
		"used_quota":    user.UsedQuota,
		"request_count": user.RequestCount,
		"group_name":    user.GroupName,
		"created_at":    user.CreatedAt,
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	if err := h.authUsecase.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "password updated successfully"})
}

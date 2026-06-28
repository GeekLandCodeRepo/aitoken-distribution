package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/resp"
	"llm-gateway/internal/user/usecase"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	search := r.URL.Query().Get("search")

	var status, role *int
	if s := r.URL.Query().Get("status"); s != "" {
		val, _ := strconv.Atoi(s)
		status = &val
	}
	if r := r.URL.Query().Get("role"); r != "" {
		val, _ := strconv.Atoi(r)
		role = &val
	}

	users, total, err := h.userUsecase.ListUsers(r.Context(), usecase.ListUsersRequest{
		Page:   page,
		Size:   size,
		Search: search,
		Status: status,
		Role:   role,
	})
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Paginated(w, total, page, size, users)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	user, err := h.userUsecase.CreateUser(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Created(w, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	user, err := h.userUsecase.GetUser(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	user, err := h.userUsecase.UpdateUser(r.Context(), id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, user)
}

func (h *UserHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	balanceBefore, balanceAfter, err := h.userUsecase.TopUp(r.Context(), id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]interface{}{
		"user_id":        id,
		"balance_before": balanceBefore,
		"balance_after":  balanceAfter,
		"amount":         req.Amount,
	})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.userUsecase.ResetPassword(r.Context(), id); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"default_password": usecase.DefaultResetPassword})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	if err := h.userUsecase.DeleteUser(r.Context(), id); err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}

	resp.Success(w, map[string]string{"message": "user deleted successfully"})
}

func (h *UserHandler) ListUserChannels(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	options, err := h.userUsecase.ListUserChannelOptions(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, options)
}

func (h *UserHandler) ReplaceUserChannels(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		resp.Error(w, errcode.ErrInvalidUserID)
		return
	}

	var req usecase.ReplaceUserChannelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	options, err := h.userUsecase.ReplaceUserChannels(r.Context(), id, req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, options)
}

func (h *UserHandler) ListChannelTemplate(w http.ResponseWriter, r *http.Request) {
	options, err := h.userUsecase.ListDefaultUserChannelOptions(r.Context())
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, options)
}

func (h *UserHandler) SaveChannelTemplate(w http.ResponseWriter, r *http.Request) {
	var req usecase.ReplaceUserChannelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	options, err := h.userUsecase.SaveDefaultUserChannels(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, options)
}

func (h *UserHandler) ApplyChannelTemplateToAllUsers(w http.ResponseWriter, r *http.Request) {
	affected, err := h.userUsecase.ApplyDefaultChannelsToAllUsers(r.Context())
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, map[string]interface{}{"affected": affected})
}

func (h *UserHandler) GetInviteSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.userUsecase.GetInviteSettings(r.Context())
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, settings)
}

func (h *UserHandler) UpdateInviteSettings(w http.ResponseWriter, r *http.Request) {
	var req usecase.InviteSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	settings, err := h.userUsecase.UpdateInviteSettings(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Success(w, settings)
}

func (h *UserHandler) CreateInviteCode(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateInviteCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error(w, errcode.ErrInvalidBody)
		return
	}

	code, err := h.userUsecase.CreateInviteCode(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			resp.Error(w, appErr)
		} else {
			resp.Error(w, errcode.ErrInternal)
		}
		return
	}
	resp.Created(w, code)
}

func (h *UserHandler) ListInviteCodes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	codes, total, err := h.userUsecase.ListInviteCodes(r.Context(), page, size)
	if err != nil {
		resp.Error(w, errcode.ErrInternal)
		return
	}
	resp.Paginated(w, total, page, size, codes)
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"goAuth/internals/helpers"
	"goAuth/internals/logger"
	"goAuth/internals/middlewares"
	"goAuth/internals/model"
	"goAuth/internals/service"
	"goAuth/internals/service/impl"
	"goAuth/internals/types"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) getService(r *http.Request) service.AuthService {
	pc := middlewares.GetProjectContext(r.Context())
	if pc == nil {
		return nil
	}
	ctx, ok := pc.(*types.ProjectContext)
	if !ok || ctx == nil || ctx.Repo == nil {
		return nil
	}
	repo, ok := ctx.Repo.(types.UserRepository)
	if !ok {
		return nil
	}
	var otpRepo types.OTPRepository
	if ctx.OTP && ctx.OTPRepo != nil {
		otpRepo, _ = ctx.OTPRepo.(types.OTPRepository)
	}
	return impl.NewAuthService(repo, otpRepo, ctx.OTP)
}

func (h *AuthHandler) getProjectContext(r *http.Request) *types.ProjectContext {
	pc := middlewares.GetProjectContext(r.Context())
	if pc == nil {
		return nil
	}
	ctx, _ := pc.(*types.ProjectContext)
	return ctx
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Logf(ctx, "signup", "-", "request received")
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		logger.Logf(ctx, "signup", "-", "service is nil – project context missing or invalid")
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	var req model.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logf(ctx, "signup", "-", "failed to decode request body: %v", err)
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}

	emailStr := "-"
	if req.Email != "" {
		emailStr = req.Email
	}

	logger.Logf(ctx, "signup", emailStr, "parsed request – email=%q phone=%q", req.Email, req.Phone)
	user, err := svc.Signup(ctx, &req)
	if err != nil {
		logger.Logf(ctx, "signup", emailStr, "service error: %v", err)
		h.writeServiceError(ctx, w, err, "signup", emailStr)
		return
	}
	data := map[string]interface{}{"user": userToResponse(user)}
	if pc := h.getProjectContext(r); pc != nil && pc.OTP {
		data["message"] = "otp_sent"
	}
	logger.Logf(ctx, "signup", emailStr, "user created successfully – id=%s", user.ID)
	helpers.WriteJSON(w, http.StatusCreated, model.DataResponse{Data: data})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Logf(ctx, "login", "-", "request received")
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		logger.Logf(ctx, "login", "-", "project context missing")
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logf(ctx, "login", "-", "failed to decode request body: %v", err)
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}
	req.EmailOrPhone = strings.TrimSpace(req.EmailOrPhone)

	identifier := req.EmailOrPhone
	if identifier == "" {
		identifier = "-"
	}

	if req.EmailOrPhone == "" {
		logger.Logf(ctx, "login", identifier, "empty email or phone")
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}

	logger.Logf(ctx, "login", identifier, "parsed request")
	user, tokens, err := svc.Login(ctx, req.EmailOrPhone, req.Password, req.UseOTP)
	if err != nil {
		if errors.Is(err, types.ErrOTPSent) {
			logger.Logf(ctx, "login", identifier, "OTP triggered successfully")
			helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: map[string]interface{}{"message": "otp_sent"}})
			return
		}
		logger.Logf(ctx, "login", identifier, "service error: %v", err)
		h.writeServiceError(ctx, w, err, "login", identifier)
		return
	}
	logger.Logf(ctx, "login", identifier, "login success for user id=%s", user.ID)
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: map[string]interface{}{
		"token":      tokens.Token,
		"expires_in": tokens.ExpiresIn,
	}})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Logf(ctx, "verify_otp", "-", "request received")
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	var req model.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logf(ctx, "verify_otp", "-", "failed to decode request body")
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}
	req.EmailOrPhone = strings.TrimSpace(req.EmailOrPhone)
	req.OTP = strings.TrimSpace(req.OTP)

	identifier := req.EmailOrPhone
	if identifier == "" {
		identifier = "-"
	}

	if req.EmailOrPhone == "" || req.OTP == "" { // ensure identifier string gets parsed right before we log
		logger.Logf(ctx, "verify_otp", identifier, "missing email/phone or OTP field")
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidOTP)
		return
	}

	logger.Logf(ctx, "verify_otp", identifier, "parsed request")
	user, tokens, err := svc.VerifyOTP(ctx, req.EmailOrPhone, req.OTP)
	if err != nil {
		logger.Logf(ctx, "verify_otp", identifier, "service error: %v", err)
		h.writeServiceError(ctx, w, err, "verify_otp", identifier)
		return
	}
	logger.Logf(ctx, "verify_otp", identifier, "OTP verified for user id=%s", user.ID)
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: map[string]interface{}{
		"token":      tokens.Token,
		"expires_in": tokens.ExpiresIn,
	}})
}

func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middlewares.GetClaims(ctx)

	if claims == nil || claims.UserID == "" {
		logger.Logf(ctx, "verify_token", "-", "unauthorized access")
		helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
		return
	}

	logger.Logf(ctx, "verify_token", claims.UserID, "token successfully verified")
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{
		Data: map[string]interface{}{
			"user_id": claims.UserID,
		},
	})
}


func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	claims := middlewares.GetClaims(ctx)
	identifier := "-"
	if claims != nil && claims.UserID != "" {
		identifier = claims.UserID
	}
	logger.Logf(ctx, "get_user", identifier, "request received")

	if claims == nil || claims.UserID == "" {
		logger.Logf(ctx, "get_user", identifier, "unauthorized access")
		helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
		return
	}
	user, err := svc.GetByID(ctx, claims.UserID)
	if err != nil {
		logger.Logf(ctx, "get_user", identifier, "service error: %v", err)
		h.writeServiceError(ctx, w, err, "get_user", identifier)
		return
	}
	logger.Logf(ctx, "get_user", identifier, "user successfully retrieved")
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: userToResponse(user)})
}

func (h *AuthHandler) writeServiceError(ctx context.Context, w http.ResponseWriter, err error, flow, identifier string) {
	switch {
	case err == types.ErrConflict:
		helpers.WriteError(w, http.StatusConflict, types.MsgConflict)
	case err == types.ErrNotFound:
		helpers.WriteError(w, http.StatusNotFound, types.MsgNotFound)
	case err == types.ErrUnauthorized, err == types.ErrInvalidPassword:
		helpers.WriteError(w, http.StatusUnauthorized, types.MsgInvalidPassword)
	case err == types.ErrInvalidToken:
		helpers.WriteError(w, http.StatusUnauthorized, types.MsgInvalidToken)
	case err == types.ErrValidation:
		helpers.WriteError(w, http.StatusBadRequest, types.MsgPasswordTooWeak)
	case err == types.ErrInvalidOTP:
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidOTP)
	default:
		logger.Logf(ctx, flow, identifier, "unhandled service error: %v", err)
		helpers.WriteError(w, http.StatusInternalServerError, types.MsgBadRequest)
	}
}

func userToResponse(u *model.User) model.UserResponse {
	return model.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

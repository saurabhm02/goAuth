package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"goAuth/internals/helpers"
	"goAuth/internals/model"
	"goAuth/internals/middlewares"
	"goAuth/internals/repository"
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
	repo, ok := ctx.Repo.(repository.UserRepository)
	if !ok {
		return nil
	}
	return impl.NewAuthService(repo)
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	var req model.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}
	user, err := svc.Signup(&req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, model.DataResponse{Data: userToResponse(user)})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}
	req.EmailOrPhone = strings.TrimSpace(req.EmailOrPhone)
	if req.EmailOrPhone == "" {
		helpers.WriteError(w, http.StatusBadRequest, types.MsgInvalidRequest)
		return
	}
	user, tokens, err := svc.Login(req.EmailOrPhone, req.Password)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: map[string]interface{}{
		"user":   userToResponse(user),
		"token":  tokens.Token,
		"expires_in": tokens.ExpiresIn,
	}})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteError(w, http.StatusMethodNotAllowed, types.MsgBadRequest)
		return
	}
	svc := h.getService(r)
	if svc == nil {
		helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
		return
	}
	claims := middlewares.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
		return
	}
	user, err := svc.GetByID(claims.UserID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, model.DataResponse{Data: userToResponse(user)})
}

func (h *AuthHandler) writeServiceError(w http.ResponseWriter, err error) {
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
	default:
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

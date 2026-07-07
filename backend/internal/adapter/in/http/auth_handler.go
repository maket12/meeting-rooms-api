package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	httpdto "github.com/maket12/meeting-rooms-api/internal/adapter/in/http/dto"
	"github.com/maket12/meeting-rooms-api/internal/adapter/in/http/mapper"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
)

type AuthHandler struct {
	BaseHandler
	dummyLoginUC *usecase.DummyLoginUC
	registerUC   *usecase.RegisterUC
	loginUC      *usecase.LoginUC
}

func NewAuthHandler(
	log *slog.Logger,
	dummyLoginUC *usecase.DummyLoginUC,
	registerUC *usecase.RegisterUC,
	loginUC *usecase.LoginUC,
) *AuthHandler {
	return &AuthHandler{
		BaseHandler:  NewBaseHandler(log),
		dummyLoginUC: dummyLoginUC,
		registerUC:   registerUC,
		loginUC:      loginUC,
	}
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req httpdto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	out, err := h.dummyLoginUC.Execute(
		r.Context(), mapper.MapRequestToDummyLogin(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to login dummy")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapDummyLoginToResponse(out))
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req httpdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	out, err := h.registerUC.Execute(
		r.Context(), mapper.MapRequestToRegister(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to register")
		return
	}

	h.log.InfoContext(r.Context(), "created user",
		slog.String("id", out.User.ID.String()),
	)

	h.respond(w, http.StatusCreated, mapper.MapRegisterToResponse(out))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req httpdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	out, err := h.loginUC.Execute(
		r.Context(), mapper.MapRequestToLogin(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to register")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapLoginToResponse(out))
}

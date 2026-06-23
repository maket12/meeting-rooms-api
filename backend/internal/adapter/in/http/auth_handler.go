package http

import (
	httpdto "backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/mapper"
	"backend/internal/app/usecase"
	pkgerrs "backend/pkg/errs"
	"encoding/json"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	log          *slog.Logger
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
		log:          log,
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

func (h *AuthHandler) handleError(w http.ResponseWriter, r *http.Request, err error, logMsg string) {
	outErr := mapper.HttpError(err)
	h.log.ErrorContext(r.Context(), logMsg,
		slog.Int("status", outErr.Code),
		slog.String("public_msg", outErr.Message),
		slog.Any("cause", outErr.Reason),
	)
	http.Error(w, outErr.Message, outErr.Code)
}

func (h *AuthHandler) respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

package http

import (
	httpdto "MeetingRoomsAPI/internal/adapter/in/http/dto"
	"MeetingRoomsAPI/internal/adapter/in/http/mapper"
	"MeetingRoomsAPI/internal/app/usecase"
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	out, err := h.dummyLoginUC.Execute(
		r.Context(), mapper.MapRequestToDummyLogin(req),
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to login dummy",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mapper.MapDummyLoginToResponse(out))
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req httpdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	out, err := h.registerUC.Execute(
		r.Context(), mapper.MapRequestToRegister(req),
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to register",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	h.log.InfoContext(r.Context(), "created user",
		slog.String("id", out.User.ID.String()),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(mapper.MapRegisterToResponse(out))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req httpdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	out, err := h.loginUC.Execute(
		r.Context(), mapper.MapRequestToLogin(req),
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to register",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mapper.MapLoginToResponse(out))
}

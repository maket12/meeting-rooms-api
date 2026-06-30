package http

import (
	"backend/internal/adapter/in/http/mapper"
	"encoding/json"
	"log/slog"
	"net/http"
)

type BaseHandler struct{ log *slog.Logger }

func NewBaseHandler(log *slog.Logger) BaseHandler {
	return BaseHandler{log: log}
}

func (h *BaseHandler) handleError(w http.ResponseWriter, r *http.Request, err error, logMsg string) {
	outErr := mapper.HttpError(err)
	h.log.ErrorContext(r.Context(), logMsg,
		slog.Int("status", outErr.Code),
		slog.String("public_msg", outErr.Message),
		slog.Any("cause", outErr.Reason),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(outErr.Code)

	response := map[string]string{"error": outErr.Message}
	_ = json.NewEncoder(w).Encode(response)
}

func (h *BaseHandler) respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

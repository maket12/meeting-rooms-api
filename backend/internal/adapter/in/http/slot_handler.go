package http

import (
	httpdto "backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/mapper"
	"backend/internal/app/usecase"
	pkgerrs "backend/pkg/errs"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SlotHandler struct {
	log         *slog.Logger
	listSlotsUC *usecase.ListSlotsUC
}

func NewSlotHandler(
	log *slog.Logger,
	listSlotsUC *usecase.ListSlotsUC,
) *SlotHandler {
	return &SlotHandler{
		log:         log,
		listSlotsUC: listSlotsUC,
	}
}

func (h *SlotHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("roomId")
	dateStr := r.URL.Query().Get("date")

	_, err := uuid.Parse(roomIdStr) // Validation of room id
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
		return
	}

	_, err = time.Parse(time.RFC3339, dateStr) // Validation of date
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidDate, "date is invalid")
	}

	req := httpdto.ListSlotsRequest{
		RoomID: roomIdStr,
		Date:   dateStr,
	}

	out, err := h.listSlotsUC.Execute(
		r.Context(), mapper.MapRequestToListSlots(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to get a list of slots")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapListSlotsToResponse(out))
}

func (h *SlotHandler) handleError(w http.ResponseWriter, r *http.Request, err error, logMsg string) {
	outErr := mapper.HttpError(err)
	h.log.ErrorContext(r.Context(), logMsg,
		slog.Int("status", outErr.Code),
		slog.String("public_msg", outErr.Message),
		slog.Any("cause", outErr.Reason),
	)
	http.Error(w, outErr.Message, outErr.Code)
}

func (h *SlotHandler) respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

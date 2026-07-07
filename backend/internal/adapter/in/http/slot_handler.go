package http

import (
	"log/slog"
	"net/http"
	"time"

	httpdto "github.com/maket12/meeting-rooms-api/internal/adapter/in/http/dto"
	"github.com/maket12/meeting-rooms-api/internal/adapter/in/http/mapper"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/google/uuid"
)

type SlotHandler struct {
	BaseHandler
	listSlotsUC *usecase.ListSlotsUC
}

func NewSlotHandler(
	log *slog.Logger,
	listSlotsUC *usecase.ListSlotsUC,
) *SlotHandler {
	return &SlotHandler{
		BaseHandler: NewBaseHandler(log),
		listSlotsUC: listSlotsUC,
	}
}

func (h *SlotHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("id")
	dateStr := r.URL.Query().Get("date")

	_, err := uuid.Parse(roomIdStr) // Validation of room id
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
		return
	}

	_, err = time.Parse(time.DateOnly, dateStr) // Validation of date
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidDate, "date is invalid")
		return
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

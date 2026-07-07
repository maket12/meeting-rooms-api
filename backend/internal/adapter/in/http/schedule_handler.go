package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	httpdto "github.com/maket12/meeting-rooms-api/internal/adapter/in/http/dto"
	"github.com/maket12/meeting-rooms-api/internal/adapter/in/http/mapper"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/google/uuid"
)

type ScheduleHandler struct {
	BaseHandler
	createScheduleUC *usecase.CreateScheduleUC
}

func NewScheduleHandler(
	log *slog.Logger,
	createScheduleUC *usecase.CreateScheduleUC,
) *ScheduleHandler {
	return &ScheduleHandler{
		BaseHandler:      NewBaseHandler(log),
		createScheduleUC: createScheduleUC,
	}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("id")
	_, err := uuid.Parse(roomIdStr) // Validation of room id
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
		return
	}

	var req httpdto.CreateScheduleRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	req.RoomID = roomIdStr

	out, err := h.createScheduleUC.Execute(
		r.Context(), mapper.MapRequestToCreateSchedule(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to create a schedule")
		return
	}

	h.log.InfoContext(r.Context(), "created schedule",
		slog.String("id", out.Schedule.ID.String()),
	)

	h.respond(w, http.StatusCreated, mapper.MapCreateScheduleToResponse(out))
}

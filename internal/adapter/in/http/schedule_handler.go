package http

import (
	httpdto "MeetingRoomsAPI/internal/adapter/in/http/dto"
	"MeetingRoomsAPI/internal/adapter/in/http/mapper"
	"MeetingRoomsAPI/internal/app/usecase"
	"encoding/json"
	"log/slog"
	"net/http"
)

type ScheduleHandler struct {
	log              *slog.Logger
	createScheduleUC *usecase.CreateScheduleUC
}

func NewScheduleHandler(
	log *slog.Logger,
	createScheduleUC *usecase.CreateScheduleUC,
) *ScheduleHandler {
	return &ScheduleHandler{
		log:              log,
		createScheduleUC: createScheduleUC,
	}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("roomId")

	var req httpdto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.RoomID = roomIdStr

	out, err := h.createScheduleUC.Execute(
		r.Context(), mapper.MapRequestToCreateSchedule(req),
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to create a schedule",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	h.log.InfoContext(r.Context(), "created schedule",
		slog.String("id", out.Schedule.ID.String()),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(mapper.MapCreateScheduleToResponse(out))
}

package http

import (
	httpdto "MeetingRoomsAPI/internal/adapter/in/http/dto"
	"MeetingRoomsAPI/internal/adapter/in/http/mapper"

	"MeetingRoomsAPI/internal/app/usecase"
	"encoding/json"
	"log/slog"
	"net/http"
)

type RoomHandler struct {
	log          *slog.Logger
	createRoomUC *usecase.CreateRoomUC
	listRoomsUC  *usecase.ListRoomsUC
}

func NewRoomHandler(
	log *slog.Logger,
	createRoomUC *usecase.CreateRoomUC,
	listRoomsUC *usecase.ListRoomsUC,
) *RoomHandler {
	return &RoomHandler{
		log:          log,
		createRoomUC: createRoomUC,
		listRoomsUC:  listRoomsUC,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req httpdto.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	out, err := h.createRoomUC.Execute(
		r.Context(), mapper.MapRequestToCreateRoom(req),
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to create a room",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	h.log.InfoContext(r.Context(), "created room",
		slog.String("id", out.ID.String()),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(mapper.MapCreateRoomToResponse(out))
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	out, err := h.listRoomsUC.Execute(r.Context())
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to get rooms list",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mapper.MapListRoomsToResponse(out))
}

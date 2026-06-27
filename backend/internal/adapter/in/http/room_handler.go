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

type RoomHandler struct {
	BaseHandler
	createRoomUC *usecase.CreateRoomUC
	listRoomsUC  *usecase.ListRoomsUC
}

func NewRoomHandler(
	log *slog.Logger,
	createRoomUC *usecase.CreateRoomUC,
	listRoomsUC *usecase.ListRoomsUC,
) *RoomHandler {
	return &RoomHandler{
		BaseHandler:  NewBaseHandler(log),
		createRoomUC: createRoomUC,
		listRoomsUC:  listRoomsUC,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req httpdto.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	out, err := h.createRoomUC.Execute(
		r.Context(), mapper.MapRequestToCreateRoom(req),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to create a room")
		return
	}

	h.log.InfoContext(r.Context(), "created room",
		slog.String("id", out.Room.ID.String()),
	)

	h.respond(w, http.StatusCreated, mapper.MapCreateRoomToResponse(out))
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	out, err := h.listRoomsUC.Execute(r.Context())
	if err != nil {
		h.handleError(w, r, err, "failed to get rooms list")
		return
	}
	h.respond(w, http.StatusOK, mapper.MapListRoomsToResponse(out))
}

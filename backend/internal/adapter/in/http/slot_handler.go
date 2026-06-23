package http

import (
	httpdto "backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/mapper"
	"backend/internal/app/usecase"
	"encoding/json"
	"log/slog"
	"net/http"
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
	roomId := r.PathValue("roomId")
	date := r.URL.Query().Get("date")

	if date == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}

	req := httpdto.ListSlotsRequest{
		RoomID: roomId,
		Date:   date,
	}

	mappedReq, err := mapper.MapRequestToListSlots(req)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	out, err := h.listSlotsUC.Execute(
		r.Context(), mappedReq,
	)
	if err != nil {
		status, msg, internalErr := mapper.HttpError(err)
		h.log.ErrorContext(r.Context(), "failed to get a list of slots",
			slog.Int("status", status),
			slog.String("public_msg", msg),
			slog.Any("cause", internalErr),
		)
		http.Error(w, msg, status)
		return
	}

	h.respond(w, http.StatusOK, mapper.MapListSlotsToResponse(out))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mapper.MapListSlotsToResponse(out))
}

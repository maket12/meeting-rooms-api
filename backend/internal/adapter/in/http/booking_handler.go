package http

import (
	httpdto "backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/mapper"
	ucdto "backend/internal/app/dto"
	"backend/internal/app/usecase"
	pkgerrs "backend/pkg/errs"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type BookingHandler struct {
	BaseHandler
	createBookingUC  *usecase.CreateBookingUC
	cancelBookingUC  *usecase.CancelBookingUC
	listMyBookingsUC *usecase.ListMyBookingsUC
	listBookingsUC   *usecase.ListBookingsUC
}

func NewBookingHandler(
	log *slog.Logger,
	createBookingUC *usecase.CreateBookingUC,
	cancelBookingUC *usecase.CancelBookingUC,
	listMyBookingsUC *usecase.ListMyBookingsUC,
	listBookingsUC *usecase.ListBookingsUC,
) *BookingHandler {
	return &BookingHandler{
		BaseHandler:      NewBaseHandler(log),
		createBookingUC:  createBookingUC,
		cancelBookingUC:  cancelBookingUC,
		listMyBookingsUC: listMyBookingsUC,
		listBookingsUC:   listBookingsUC,
	}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok { // Validation of user id (e.g. authorization)
		h.handleError(w, r, pkgerrs.ErrInvalidUserID, "unauthorized: user id not found in context")
		return
	}

	var req httpdto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidJSON, "invalid json")
		return
	}

	// Validation of slot id
	if _, err := uuid.Parse(req.SlotID); err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
		return
	}

	out, err := h.createBookingUC.Execute(
		r.Context(),
		mapper.MapRequestToCreateBooking(req, userID),
	)
	if err != nil {
		h.handleError(w, r, err, "failed to create booking")
		return
	}

	h.log.InfoContext(r.Context(), "created booking", slog.String("id", out.Booking.ID.String()))

	h.respond(w, http.StatusCreated, mapper.MapCreateBookingToResponse(out))
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok { // Validation of user id (e.g. authorization)
		h.handleError(w, r, pkgerrs.ErrInvalidUserID, "unauthorized: user id not found in context")
		return
	}

	bookingIDStr := r.PathValue("id")
	bookingID, err := uuid.Parse(bookingIDStr) // Validation of booking id
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
		return
	}

	out, err := h.cancelBookingUC.Execute(r.Context(), ucdto.CancelBookingInput{
		BookingID:   bookingID,
		RequestorID: userID,
	})
	if err != nil {
		h.handleError(w, r, err, "failed to cancel booking")
		return
	}

	h.log.InfoContext(r.Context(), "cancelled booking", slog.String("id", bookingIDStr))

	h.respond(w, http.StatusOK, mapper.MapCancelBookingToResponse(out))
}

func (h *BookingHandler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok { // Validation of user id (e.g. authorization)
		h.handleError(w, r, pkgerrs.ErrInvalidUserID, "unauthorized: user id not found in context")
		return
	}

	out, err := h.listMyBookingsUC.Execute(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err, "failed to list my bookings")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapListMyBookingsToResponse(out))
}

func (h *BookingHandler) ListAllBookings(w http.ResponseWriter, r *http.Request) {
	var page, pageSize int

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	if len(pageStr) != 0 {
		pageInt, err := strconv.Atoi(pageStr)
		if err != nil {
			h.handleError(w, r, pkgerrs.ErrInvalidPage, "invalid page value")
			return
		}
		page = pageInt
	}

	if len(pageSizeStr) != 0 {
		pageSizeInt, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			h.handleError(w, r, pkgerrs.ErrInvalidPage, "invalid page size value")
			return
		}
		pageSize = pageSizeInt
	}

	out, err := h.listBookingsUC.Execute(r.Context(), ucdto.ListBookingsInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.handleError(w, r, err, "failed to list all bookings")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapListAllBookingsToResponse(out))
}

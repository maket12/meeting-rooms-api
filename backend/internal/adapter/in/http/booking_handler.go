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
	log              *slog.Logger
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
		log:              log,
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

	h.respond(w, http.StatusCreated, mapper.MapBookingToResponse(out.Booking))
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok { // Validation of user id (e.g. authorization)
		h.handleError(w, r, pkgerrs.ErrInvalidUserID, "unauthorized: user id not found in context")
		return
	}

	bookingIDStr := r.PathValue("bookingId")
	bookingID, err := uuid.Parse(bookingIDStr) // Validation of booking id
	if err != nil {
		h.handleError(w, r, pkgerrs.ErrInvalidIdentifier, "failed to parse uuid")
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

	h.respond(w, http.StatusOK, mapper.MapBookingToResponse(out.Booking))
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	out, err := h.listBookingsUC.Execute(r.Context(), ucdto.ListBookingsInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.handleError(w, r, err, "failed to list all bookings")
		return
	}

	h.respond(w, http.StatusOK, mapper.MapListToResponse(out))
}

func (h *BookingHandler) handleError(w http.ResponseWriter, r *http.Request, err error, logMsg string) {
	outErr := mapper.HttpError(err)
	h.log.ErrorContext(r.Context(), logMsg,
		slog.Int("status", outErr.Code),
		slog.String("public_msg", outErr.Message),
		slog.Any("cause", outErr.Reason),
	)
	http.Error(w, outErr.Message, outErr.Code)
}

func (h *BookingHandler) respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

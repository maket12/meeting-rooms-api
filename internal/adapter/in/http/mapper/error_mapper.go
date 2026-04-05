package mapper

import (
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"errors"
	"net/http"
)

//func HttpError(err error) (int, string, error) {
//	var w *ucerrs.WrappedError
//	if errors.As(err, &w) {
//		switch {
//		case errors.Is(err, ucerrs.ErrCreateUserDB),
//			errors.Is(err, ucerrs.ErrGetUserByIDDB),
//			errors.Is(err, ucerrs.ErrGetUserByEmailDB),
//			errors.Is(err, ucerrs.ErrCreateRoomDB),
//			errors.Is(err, ucerrs.ErrListRoomsDB),
//			errors.Is(err, ucerrs.ErrCreateScheduleDB),
//			errors.Is(err, ucerrs.ErrGetScheduleDB),
//			errors.Is(err, ucerrs.ErrCreateSlotsDB),
//			errors.Is(err, ucerrs.ErrGetSlotDB),
//			errors.Is(err, ucerrs.ErrListSlotsDB),
//			errors.Is(err, ucerrs.ErrCreateBookingDB),
//			errors.Is(err, ucerrs.ErrGetBookingDB),
//			errors.Is(err, ucerrs.ErrUpdateBookingStatusDB),
//			errors.Is(err, ucerrs.ErrListBookingsDB),
//			errors.Is(err, ucerrs.ErrListMyBookingsDB),
//			errors.Is(err, ucerrs.ErrHashPassword),
//			errors.Is(err, ucerrs.ErrGenerateToken),
//			errors.Is(err, ucerrs.ErrCreateMeeting):
//			return http.StatusInternalServerError, w.Public.Error(), w.Reason
//
//		case errors.Is(err, ucerrs.ErrInvalidInput):
//			return http.StatusBadRequest, w.Public.Error(), w.Reason
//
//		default:
//			return http.StatusInternalServerError, "internal error", w.Reason
//		}
//	}
//
//	switch {
//	case errors.Is(err, ucerrs.ErrInvalidCredentials):
//		return http.StatusUnauthorized, err.Error(), nil
//
//	case errors.Is(err, ucerrs.ErrCannotCancelBooking):
//		return http.StatusForbidden, err.Error(), nil
//
//	case errors.Is(err, ucerrs.ErrRoomNotFound),
//		errors.Is(err, ucerrs.ErrScheduleNotFound),
//		errors.Is(err, ucerrs.ErrSlotNotFound),
//		errors.Is(err, ucerrs.ErrBookingNotFound):
//		return http.StatusNotFound, err.Error(), nil
//
//	case errors.Is(err, ucerrs.ErrUserAlreadyExists),
//		errors.Is(err, ucerrs.ErrCannotCreateBooking):
//		return http.StatusBadRequest, err.Error(), nil
//
//	case errors.Is(err, ucerrs.ErrScheduleAlreadyExists):
//		return http.StatusConflict, err.Error(), nil
//	}
//
//	return http.StatusInternalServerError, "internal error", err
//}

func MapAppErrorToAPI(err error) (int, string, string) {
	if errors.Is(err, ucerrs.ErrRoomNotFound) {
		return http.StatusNotFound, "ROOM_NOT_FOUND", "room not found"
	}
	if errors.Is(err, ucerrs.ErrSlotNotFound) {
		return http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found"
	}
	if errors.Is(err, ucerrs.ErrBookingNotFound) {
		return http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found"
	}

	//if errors.Is(err, ucerrs.ErrSlotAlreadyBooked) {
	//	return http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked"
	//}
	if errors.Is(err, ucerrs.ErrScheduleAlreadyExists) {
		return http.StatusConflict, "SCHEDULE_EXISTS", "schedule already exists"
	}

	if errors.Is(err, ucerrs.ErrInvalidCredentials) {
		return http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials"
	}
	if errors.Is(err, ucerrs.ErrForbidden) || errors.Is(err, ucerrs.ErrCannotCancelBooking) {
		return http.StatusForbidden, "FORBIDDEN", "access denied"
	}
	if errors.Is(err, ucerrs.ErrPastBooking) || errors.Is(err, ucerrs.ErrInvalidInput) {
		return http.StatusBadRequest, "INVALID_REQUEST", "invalid request parameters"
	}

	return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"
}

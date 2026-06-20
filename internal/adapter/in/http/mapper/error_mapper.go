package mapper

import (
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"errors"
	"net/http"
)

func HttpError(err error) *pkgerrs.OutErr {
	var w *ucerrs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(err, ucerrs.ErrCreateUserDB),
			errors.Is(err, ucerrs.ErrGetUserByIDDB),
			errors.Is(err, ucerrs.ErrGetUserByEmailDB),
			errors.Is(err, ucerrs.ErrCreateRoomDB),
			errors.Is(err, ucerrs.ErrListRoomsDB),
			errors.Is(err, ucerrs.ErrCreateScheduleDB),
			errors.Is(err, ucerrs.ErrGetScheduleDB),
			errors.Is(err, ucerrs.ErrCreateSlotsDB),
			errors.Is(err, ucerrs.ErrGetSlotDB),
			errors.Is(err, ucerrs.ErrListSlotsDB),
			errors.Is(err, ucerrs.ErrCreateBookingDB),
			errors.Is(err, ucerrs.ErrGetBookingDB),
			errors.Is(err, ucerrs.ErrUpdateBookingStatusDB),
			errors.Is(err, ucerrs.ErrListBookingsDB),
			errors.Is(err, ucerrs.ErrListMyBookingsDB),
			errors.Is(err, ucerrs.ErrHashPassword),
			errors.Is(err, ucerrs.ErrGenerateToken),
			errors.Is(err, ucerrs.ErrCreateMeeting):
			return http.StatusInternalServerError, w.Public.Error(), w.Reason

		case errors.Is(err, ucerrs.ErrInvalidInput):
			return pkgerrs.NewOutError(
				http.StatusBadRequest, "INVALID_REQUEST",
				"invalid_request", w.Reason,
			)

		default:
			return http.StatusInternalServerError, "internal error", w.Reason
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrInvalidCredentials):
		return http.StatusUnauthorized, err.Error(), nil

	case errors.Is(err, ucerrs.ErrCannotCancelBooking):
		return http.StatusForbidden, err.Error(), nil

	case errors.Is(err, ucerrs.ErrRoomNotFound),
		errors.Is(err, ucerrs.ErrScheduleNotFound),
		errors.Is(err, ucerrs.ErrSlotNotFound),
		errors.Is(err, ucerrs.ErrBookingNotFound):
		return http.StatusNotFound, err.Error(), nil

	case errors.Is(err, ucerrs.ErrUserAlreadyExists),
		errors.Is(err, ucerrs.ErrCannotCreateBooking):
		return pkgerrs.NewOutError(
			http.StatusBadRequest, "INVALID_REQUEST",
			"invalid_request", err,
		)

	case errors.Is(err, ucerrs.ErrScheduleAlreadyExists):
		return http.StatusConflict, err.Error(), nil
	}

	return http.StatusInternalServerError, "internal error", err
}

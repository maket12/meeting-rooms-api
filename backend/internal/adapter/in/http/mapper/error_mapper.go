package mapper

import (
	"backend/internal/app/errs"
	pkgerrs "backend/pkg/errs"
	"errors"
	"net/http"
)

func HttpError(err error) *pkgerrs.OutErr {
	if err == nil {
		return nil
	}

	if errors.Is(err, pkgerrs.ErrInvalidUserID) {
		return pkgerrs.NewOutError(
			http.StatusUnauthorized,
			err.Error(),
			err,
		)
	}

	switch {
	case errors.Is(err, pkgerrs.ErrInvalidJSON),
		errors.Is(err, pkgerrs.ErrInvalidIdentifier),
		errors.Is(err, pkgerrs.ErrValueIsInvalid):
		return pkgerrs.NewOutError(
			http.StatusBadRequest,
			err.Error(),
			err,
		)
	}

	var w *errs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(err, errs.ErrCreateUserDB),
			errors.Is(err, errs.ErrGetUserByIDDB),
			errors.Is(err, errs.ErrGetUserByEmailDB),
			errors.Is(err, errs.ErrCreateRoomDB),
			errors.Is(err, errs.ErrListRoomsDB),
			errors.Is(err, errs.ErrCreateScheduleDB),
			errors.Is(err, errs.ErrGetScheduleDB),
			errors.Is(err, errs.ErrCreateSlotsDB),
			errors.Is(err, errs.ErrGetSlotDB),
			errors.Is(err, errs.ErrListSlotsDB),
			errors.Is(err, errs.ErrCreateBookingDB),
			errors.Is(err, errs.ErrGetBookingDB),
			errors.Is(err, errs.ErrUpdateBookingStatusDB),
			errors.Is(err, errs.ErrListBookingsDB),
			errors.Is(err, errs.ErrListMyBookingsDB),
			errors.Is(err, errs.ErrHashPassword),
			errors.Is(err, errs.ErrGenerateToken),
			errors.Is(err, errs.ErrCreateMeeting):
			return pkgerrs.NewOutError(
				http.StatusInternalServerError,
				w.Public.Error(),
				w.Reason,
			)

		case errors.Is(err, errs.ErrInvalidInput):
			return pkgerrs.NewOutError(
				http.StatusBadRequest,
				w.Public.Error(),
				w.Reason,
			)

		default:
			return pkgerrs.NewOutError(
				http.StatusInternalServerError,
				"internal error",
				w.Reason,
			)
		}
	}

	switch {
	case errors.Is(err, errs.ErrInvalidCredentials):
		return pkgerrs.NewOutError(
			http.StatusUnauthorized,
			err.Error(),
			nil,
		)

	case errors.Is(err, errs.ErrCannotCancelBooking):
		return pkgerrs.NewOutError(
			http.StatusForbidden,
			err.Error(),
			nil,
		)

	case errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrRoomNotFound),
		errors.Is(err, errs.ErrScheduleNotFound),
		errors.Is(err, errs.ErrSlotNotFound),
		errors.Is(err, errs.ErrBookingNotFound):
		return pkgerrs.NewOutError(
			http.StatusNotFound,
			err.Error(),
			nil,
		)

	case errors.Is(err, errs.ErrCannotCreateBooking),
		errors.Is(err, errs.ErrUserAlreadyExists),
		errors.Is(err, errs.ErrScheduleAlreadyExists):
		return pkgerrs.NewOutError(
			http.StatusConflict,
			err.Error(),
			nil,
		)
	}

	return pkgerrs.NewOutError(
		http.StatusInternalServerError,
		"internal error",
		nil,
	)
}

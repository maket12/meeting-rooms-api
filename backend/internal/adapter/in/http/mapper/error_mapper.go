package mapper

import (
	ucerrs "backend/internal/app/errs"
	pkgucerrs "backend/pkg/errs"
	"errors"
	"net/http"
)

func HttpError(err error) *pkgucerrs.OutErr {
	if err == nil {
		return nil
	}

	if errors.Is(err, pkgucerrs.ErrInvalidUserID) {
		return pkgucerrs.NewOutError(
			http.StatusUnauthorized,
			err.Error(),
			err,
		)
	}

	switch {
	case errors.Is(err, pkgucerrs.ErrInvalidJSON),
		errors.Is(err, pkgucerrs.ErrInvalidIdentifier),
		errors.Is(err, pkgucerrs.ErrValueIsInvalid),
		errors.Is(err, pkgucerrs.ErrInvalidUserID),
		errors.Is(err, pkgucerrs.ErrInvalidDate),
		errors.Is(err, pkgucerrs.ErrInvalidPage):
		return pkgucerrs.NewOutError(
			http.StatusBadRequest,
			err.Error(),
			err,
		)
	}

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
			errors.Is(err, ucerrs.ErrExistsForDateDB),
			errors.Is(err, ucerrs.ErrCreateBookingDB),
			errors.Is(err, ucerrs.ErrGetBookingDB),
			errors.Is(err, ucerrs.ErrUpdateBookingStatusDB),
			errors.Is(err, ucerrs.ErrListBookingsDB),
			errors.Is(err, ucerrs.ErrListMyBookingsDB),
			errors.Is(err, ucerrs.ErrHashPassword),
			errors.Is(err, ucerrs.ErrGenerateToken),
			errors.Is(err, ucerrs.ErrCreateMeeting):
			return pkgucerrs.NewOutError(
				http.StatusInternalServerError,
				w.Public.Error(),
				w.Reason,
			)

		case errors.Is(err, ucerrs.ErrInvalidInput):
			return pkgucerrs.NewOutError(
				http.StatusBadRequest,
				w.Public.Error()+": "+w.Reason.Error(),
				w.Reason,
			)

		default:
			return pkgucerrs.NewOutError(
				http.StatusInternalServerError,
				"internal error",
				w.Reason,
			)
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrInvalidCredentials):
		return pkgucerrs.NewOutError(
			http.StatusUnauthorized,
			err.Error(),
			nil,
		)

	case errors.Is(err, ucerrs.ErrCannotCancelBooking):
		return pkgucerrs.NewOutError(
			http.StatusForbidden,
			err.Error(),
			nil,
		)

	case errors.Is(err, ucerrs.ErrUserNotFound),
		errors.Is(err, ucerrs.ErrRoomNotFound),
		errors.Is(err, ucerrs.ErrScheduleNotFound),
		errors.Is(err, ucerrs.ErrSlotNotFound),
		errors.Is(err, ucerrs.ErrBookingNotFound):
		return pkgucerrs.NewOutError(
			http.StatusNotFound,
			err.Error(),
			nil,
		)

	case errors.Is(err, ucerrs.ErrCannotCreateBooking),
		errors.Is(err, ucerrs.ErrUserAlreadyExists),
		errors.Is(err, ucerrs.ErrScheduleAlreadyExists),
		errors.Is(err, ucerrs.ErrBookingAlreadyExists):
		return pkgucerrs.NewOutError(
			http.StatusConflict,
			err.Error(),
			nil,
		)
	}

	return pkgucerrs.NewOutError(
		http.StatusInternalServerError,
		"internal error",
		nil,
	)
}

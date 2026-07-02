package usecase

import (
	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
)

type CancelBookingUC struct{ booking port.BookingRepository }

func NewCancelBookingUC(booking port.BookingRepository) *CancelBookingUC {
	return &CancelBookingUC{booking: booking}
}

func (uc *CancelBookingUC) Execute(ctx context.Context, in dto.CancelBookingInput) (dto.CancelBookingOutput, error) {
	booking, err := uc.booking.Get(ctx, in.BookingID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.CancelBookingOutput{}, ucerrs.ErrBookingNotFound
		}
		return dto.CancelBookingOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetBookingDB, err,
		)
	}

	if err = booking.Cancel(in.RequestorID); err != nil {
		return dto.CancelBookingOutput{}, ucerrs.ErrCannotCancelBooking
	}

	err = uc.booking.UpdateStatus(ctx, booking.ID(), booking.Status())
	if err != nil {
		return dto.CancelBookingOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateBookingStatusDB, err,
		)
	}

	return dto.CancelBookingOutput{
		Booking: mapper.MapDomainToBookingDTO(booking),
	}, nil
}

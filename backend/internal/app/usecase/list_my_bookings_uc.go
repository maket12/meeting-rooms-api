package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
	"context"

	"github.com/google/uuid"
)

type ListMyBookingsUC struct {
	bookingRepo port.BookingRepository
}

func NewListMyBookingsUC(bookingRepo port.BookingRepository) *ListMyBookingsUC {
	return &ListMyBookingsUC{bookingRepo: bookingRepo}
}

func (uc *ListMyBookingsUC) Execute(ctx context.Context, userID uuid.UUID) (dto.ListMyBookingsOutput, error) {
	bookings, err := uc.bookingRepo.ListByUserID(ctx, userID)
	if err != nil {
		return dto.ListMyBookingsOutput{}, errs.Wrap(
			errs.ErrListMyBookingsDB, err,
		)
	}

	return mapper.MapDomainToListMyBookings(bookings), nil
}

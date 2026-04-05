package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/port"
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
		return dto.ListMyBookingsOutput{}, ucerrs.Wrap(
			ucerrs.ErrListMyBookingsDB, err,
		)
	}

	return mapper.MapDomainToListMyBookings(bookings), nil
}

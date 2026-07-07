package usecase

import (
	"context"

	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"

	"github.com/google/uuid"
)

type ListMyBookingsUC struct{ bookingRepo port.BookingRepository }

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

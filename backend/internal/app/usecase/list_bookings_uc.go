package usecase

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
)

type ListBookingsUC struct{ booking port.BookingRepository }

func NewListBookingsUC(bookingRepo port.BookingRepository) *ListBookingsUC {
	return &ListBookingsUC{booking: bookingRepo}
}

func (uc *ListBookingsUC) Execute(ctx context.Context, input dto.ListBookingsInput) (dto.ListBookingsOutput, error) {
	if input.Page < 0 {
		return dto.ListBookingsOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, errors.New("page can't be negative"),
		)
	}

	if input.PageSize < 0 {
		return dto.ListBookingsOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, errors.New("page size can't be negative"),
		)
	}

	limit := int32(input.PageSize)
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	page := input.Page
	if page == 0 {
		page = 1
	}

	offset := int32((page - 1) * int(limit))

	bookings, total, err := uc.booking.ListAll(ctx, limit, offset)
	if err != nil {
		return dto.ListBookingsOutput{}, ucerrs.Wrap(
			ucerrs.ErrListBookingsDB, err,
		)
	}

	output := dto.ListBookingsOutput{
		Bookings: mapper.MapDomainToListBookings(bookings).Bookings,
		Pagination: dto.Pagination{
			Page:     page,
			PageSize: int(limit),
			Total:    int(total),
		},
	}

	return output, nil
}

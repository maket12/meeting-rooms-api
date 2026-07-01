package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
	"context"
)

type ListBookingsUC struct {
	booking port.BookingRepository
}

func NewListBookingsUC(bookingRepo port.BookingRepository) *ListBookingsUC {
	return &ListBookingsUC{booking: bookingRepo}
}

func (uc *ListBookingsUC) Execute(ctx context.Context, input dto.ListBookingsInput) (dto.ListBookingsOutput, error) {
	limit := int32(input.PageSize)
	if limit <= 0 {
		limit = 10
	}

	page := input.Page
	if page <= 0 {
		page = 1
	}

	offset := int32((page - 1) * int(limit))

	bookings, total, err := uc.booking.ListAll(ctx, limit, offset)
	if err != nil {
		return dto.ListBookingsOutput{}, errs.Wrap(
			errs.ErrListBookingsDB, err,
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

package usecase

import (
	dto2 "backend/internal/app/dto"
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

func (uc *ListBookingsUC) Execute(ctx context.Context, input dto2.ListBookingsInput) (dto2.ListBookingsOutput, error) {
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
		return dto2.ListBookingsOutput{}, errs.Wrap(
			errs.ErrListBookingsDB, err,
		)
	}

	output := dto2.ListBookingsOutput{
		Bookings: mapper.MapDomainToListBookings(bookings).Bookings,
		Pagination: dto2.Pagination{
			Page:     page,
			PageSize: int(limit),
			Total:    int(total),
		},
	}

	return output, nil
}

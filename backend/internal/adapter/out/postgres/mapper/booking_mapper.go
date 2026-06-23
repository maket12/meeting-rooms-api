package mapper

import (
	sqlc2 "backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapBookingToSQLCCreate(booking *model.Booking) sqlc2.CreateBookingParams {
	var confLink pgtype.Text
	if booking.ConferenceLink() != nil {
		confLink = pgtype.Text{
			String: *booking.ConferenceLink(),
			Valid:  true,
		}
	}

	return sqlc2.CreateBookingParams{
		ID: pgtype.UUID{
			Bytes: booking.ID(),
			Valid: true,
		},
		SlotID: pgtype.UUID{
			Bytes: booking.SlotID(),
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: booking.UserID(),
			Valid: true,
		},
		Status:         booking.Status().String(),
		ConferenceLink: confLink,
		CreatedAt: pgtype.Timestamptz{
			Time:             booking.CreatedAt(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}
}

func MapSQLCToBooking(rawBooking sqlc2.Booking) *model.Booking {
	var confLink *string
	if rawBooking.ConferenceLink.Valid {
		confLink = &rawBooking.ConferenceLink.String
	}

	return model.RestoreBooking(
		rawBooking.ID.Bytes,
		rawBooking.SlotID.Bytes,
		rawBooking.UserID.Bytes,
		model.BookingStatus(rawBooking.Status),
		confLink,
		rawBooking.CreatedAt.Time.UTC(),
	)
}

func MapSQLCToBookingsList(rawBookings []sqlc2.Booking) []*model.Booking {
	bookings := make([]*model.Booking, len(rawBookings))
	for i := range bookings {
		mapped := MapSQLCToBooking(rawBookings[i])
		bookings[i] = mapped
	}
	return bookings
}

func MapSQLCAllToBookingsList(raw []sqlc2.ListAllBookingsRow) ([]*model.Booking, int64) {
	if len(raw) == 0 {
		return []*model.Booking{}, 0
	}

	bookings := make([]*model.Booking, len(raw))
	for i := range bookings {
		var confLink *string
		if raw[i].ConferenceLink.Valid {
			confLink = &raw[i].ConferenceLink.String
		}

		booking := model.RestoreBooking(
			raw[i].ID.Bytes,
			raw[i].SlotID.Bytes,
			raw[i].UserID.Bytes,
			model.BookingStatus(raw[i].Status),
			confLink,
			raw[i].CreatedAt.Time.UTC(),
		)
		bookings[i] = booking
	}

	total := raw[0].TotalCount

	return bookings, total
}

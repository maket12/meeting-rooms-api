package mapper_test

import (
	"backend/internal/adapter/out/postgres/mapper"
	sqlc2 "backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/model"
	"backend/pkg/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapBookingToSQLCCreate(t *testing.T) {
	booking, _ := model.NewBooking(
		uuid.New(),
		uuid.New(),
		utils.VPtr("https://yandex.telemost.ru/kai45hgt78q11"),
	)

	mapped := mapper.MapBookingToSQLCCreate(booking)

	require.True(t, mapped.ID.Valid)
	require.True(t, mapped.SlotID.Valid)
	require.True(t, mapped.UserID.Valid)
	require.True(t, mapped.CreatedAt.Valid)

	assert.Equal(t, [16]byte(booking.ID()), mapped.ID.Bytes)
	assert.Equal(t, [16]byte(booking.SlotID()), mapped.SlotID.Bytes)
	assert.Equal(t, [16]byte(booking.UserID()), mapped.UserID.Bytes)
	assert.Equal(t, booking.Status().String(), mapped.Status)
	assert.Equal(t, booking.CreatedAt(), mapped.CreatedAt.Time)
}

func TestMapSQLCToBooking(t *testing.T) {
	raw := sqlc2.Booking{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		SlotID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Status: "active",
		ConferenceLink: pgtype.Text{
			String: "https://yandex.telemost.ru/kai45hgt78q11",
			Valid:  true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}

	mapped := mapper.MapSQLCToBooking(raw)

	assert.Equal(t, raw.ID.Bytes, [16]byte(mapped.ID()))
	assert.Equal(t, raw.SlotID.Bytes, [16]byte(mapped.SlotID()))
	assert.Equal(t, raw.UserID.Bytes, [16]byte(mapped.UserID()))
	assert.Equal(t, raw.Status, mapped.Status().String())
	assert.Equal(t, raw.ConferenceLink.String, *mapped.ConferenceLink())
	assert.Equal(t, raw.CreatedAt.Time, mapped.CreatedAt())
}

func TestMapSQLCToBookingsList(t *testing.T) {
	raw := []sqlc2.Booking{
		{
			ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			SlotID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			UserID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Status:    "active",
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		},
		{
			ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			SlotID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			UserID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Status:    "cancelled",
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		},
	}

	mapped := mapper.MapSQLCToBookingsList(raw)

	require.Len(t, mapped, 2)
	assert.Equal(t, raw[0].ID.Bytes, [16]byte(mapped[0].ID()))
	assert.Equal(t, raw[1].ID.Bytes, [16]byte(mapped[1].ID()))
	assert.Equal(t, "active", mapped[0].Status().String())
	assert.Equal(t, "cancelled", mapped[1].Status().String())
}

func TestMapSQLCAllToBookingsList(t *testing.T) {
	t.Run("success_mapping", func(t *testing.T) {
		raw := []sqlc2.ListAllBookingsRow{
			{
				ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
				SlotID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
				UserID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Status:         "active",
				ConferenceLink: pgtype.Text{String: "https://link.com", Valid: true},
				CreatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
				TotalCount:     10,
			},
			{
				ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
				SlotID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
				UserID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Status:         "cancelled",
				ConferenceLink: pgtype.Text{Valid: false},
				CreatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
				TotalCount:     10,
			},
		}

		mapped, total := mapper.MapSQLCAllToBookingsList(raw)

		assert.Equal(t, int64(10), total)
		require.Len(t, mapped, 2)
		assert.Equal(t, raw[0].ID.Bytes, [16]byte(mapped[0].ID()))
		assert.Equal(t, "https://link.com", *mapped[0].ConferenceLink())
		assert.Nil(t, mapped[1].ConferenceLink())
	})

	t.Run("empty_input", func(t *testing.T) {
		mapped, total := mapper.MapSQLCAllToBookingsList([]sqlc2.ListAllBookingsRow{})

		assert.Empty(t, mapped)
		assert.Equal(t, int64(0), total)
	})
}

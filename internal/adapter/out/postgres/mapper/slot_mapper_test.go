package mapper_test

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/mapper"
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSlotsToSQLCCreateBatch(t *testing.T) {
	var (
		roomID    = uuid.New()
		startTime = time.Now().UTC()
	)

	slots := []*model.Slot{
		model.RestoreSlot(
			uuid.New(), roomID,
			startTime, startTime.Add(time.Hour),
		),
		model.RestoreSlot(
			uuid.New(), roomID,
			startTime, startTime.Add(2*time.Hour),
		),
	}

	mapped := mapper.MapSlotsToSQLCCreateBatch(slots)
	require.Len(t, mapped, len(slots))

	for i := range mapped {
		require.True(t, mapped[i].ID.Valid)
		require.True(t, mapped[i].RoomID.Valid)
		require.True(t, mapped[i].StartTime.Valid)
		require.True(t, mapped[i].EndTime.Valid)

		assert.Equal(t, [16]byte(slots[i].ID()), mapped[i].ID.Bytes)
		assert.Equal(t, [16]byte(slots[i].RoomID()), mapped[i].RoomID.Bytes)
		assert.Equal(t, slots[i].Start(), mapped[i].StartTime.Time)
		assert.Equal(t, slots[i].End(), mapped[i].EndTime.Time)
	}
}

func TestMapSQLCToSlot(t *testing.T) {
	rawSlot := sqlc.Slot{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		StartTime: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
		EndTime: pgtype.Timestamptz{
			Time:             time.Now().UTC().Add(time.Hour),
			InfinityModifier: 0,
			Valid:            true,
		},
	}

	mapped := mapper.MapSQLCToSlot(rawSlot)
	require.NotNil(t, mapped)
	assert.Equal(t, rawSlot.ID.Bytes, [16]byte(mapped.ID()))
	assert.Equal(t, rawSlot.RoomID.Bytes, [16]byte(mapped.RoomID()))
	assert.Equal(t, rawSlot.StartTime.Time, mapped.Start())
	assert.Equal(t, rawSlot.EndTime.Time, mapped.End())
}

func TestMapSQLCToSlots(t *testing.T) {
	slot1 := sqlc.Slot{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		StartTime: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
		EndTime: pgtype.Timestamptz{
			Time:             time.Now().UTC().Add(time.Hour),
			InfinityModifier: 0,
			Valid:            true,
		},
	}
	slot2 := sqlc.Slot{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		StartTime: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
		EndTime: pgtype.Timestamptz{
			Time:             time.Now().UTC().Add(time.Hour),
			InfinityModifier: 0,
			Valid:            true,
		},
	}

	rawSlots := []sqlc.Slot{slot1, slot2}

	mapped := mapper.MapSQLCToSlots(rawSlots)
	require.Len(t, mapped, len(rawSlots))

	for i := range mapped {
		assert.Equal(t, rawSlots[i].ID.Bytes, [16]byte(mapped[i].ID()))
		assert.Equal(t, rawSlots[i].RoomID.Bytes, [16]byte(mapped[i].RoomID()))
		assert.Equal(t, rawSlots[i].StartTime.Time, mapped[i].Start())
		assert.Equal(t, rawSlots[i].EndTime.Time, mapped[i].End())
	}
}

package mapper_test

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/mapper"
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMapScheduleToSQLCCreate(t *testing.T) {
	t.Parallel()

	schedule, _ := model.NewSchedule(
		uuid.New(),
		[]int{1, 3, 5},
		"10:00",
		"11:00",
	)

	mapped := mapper.MapScheduleToSQLCCreate(schedule)

	assert.Equal(t, [16]byte(schedule.ID()), mapped.ID.Bytes)
	assert.Equal(t, [16]byte(schedule.RoomID()), mapped.RoomID.Bytes)
	assert.Equal(t, len(schedule.DaysOfWeek()), len(mapped.DaysOfWeek))
	assert.Equal(t, schedule.StartTime().TotalMinutes(), mapped.StartMinutes)
	assert.Equal(t, schedule.EndTime().TotalMinutes(), mapped.EndMinutes)
}

func TestMapSQLCToSchedule(t *testing.T) {
	t.Parallel()

	rawSchedule := sqlc.Schedule{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		DaysOfWeek:   []int32{1, 3, 5},
		StartMinutes: 720,
		EndMinutes:   1230,
	}

	schedule := mapper.MapSQLCToSchedule(rawSchedule)

	assert.Equal(t, rawSchedule.ID.Bytes, [16]byte(schedule.ID()))
	assert.Equal(t, rawSchedule.RoomID.Bytes, [16]byte(schedule.RoomID()))
	assert.Equal(t, len(rawSchedule.DaysOfWeek), len(schedule.DaysOfWeek()))
	assert.Equal(t, rawSchedule.StartMinutes, schedule.StartTime().TotalMinutes())
	assert.Equal(t, rawSchedule.EndMinutes, schedule.EndTime().TotalMinutes())
}

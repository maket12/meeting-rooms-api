package mapper

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapScheduleToSQLCCreate(schedule *model.Schedule) sqlc.CreateScheduleParams {
	daysOfWeek := make([]int32, len(schedule.DaysOfWeek()))
	for i := range daysOfWeek {
		daysOfWeek[i] = int32(schedule.DaysOfWeek()[i])
	}

	return sqlc.CreateScheduleParams{
		ID: pgtype.UUID{
			Bytes: schedule.ID(),
			Valid: true,
		},
		RoomID: pgtype.UUID{
			Bytes: schedule.RoomID(),
			Valid: true,
		},
		DaysOfWeek:   daysOfWeek,
		StartMinutes: schedule.StartTime().TotalMinutes(),
		EndMinutes:   schedule.EndTime().TotalMinutes(),
	}
}

func MapSQLCToSchedule(rawSchedule sqlc.Schedule) *model.Schedule {
	daysOfWeek := make([]int, len(rawSchedule.DaysOfWeek))
	for i := range daysOfWeek {
		daysOfWeek[i] = int(rawSchedule.DaysOfWeek[i])
	}

	return model.RestoreSchedule(
		rawSchedule.ID.Bytes,
		rawSchedule.RoomID.Bytes,
		daysOfWeek,
		model.RestoreDayTime(rawSchedule.StartMinutes),
		model.RestoreDayTime(rawSchedule.EndMinutes),
	)
}

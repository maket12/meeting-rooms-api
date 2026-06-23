package mapper

import (
	sqlc2 "backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapScheduleToSQLCCreate(schedule *model.Schedule) sqlc2.CreateScheduleParams {
	daysOfWeek := make([]int32, len(schedule.DaysOfWeek()))
	for i := range daysOfWeek {
		daysOfWeek[i] = int32(schedule.DaysOfWeek()[i])
	}

	return sqlc2.CreateScheduleParams{
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

func MapSQLCToSchedule(rawSchedule sqlc2.Schedule) *model.Schedule {
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

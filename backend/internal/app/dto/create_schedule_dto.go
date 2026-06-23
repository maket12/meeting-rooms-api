package dto

import "github.com/google/uuid"

type CreateScheduleInput struct {
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type CreateScheduleOutput struct {
	Schedule Schedule
}

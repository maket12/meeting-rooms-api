package dto

import "github.com/google/uuid"

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

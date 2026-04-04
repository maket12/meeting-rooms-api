package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListSlotsInput struct {
	RoomID uuid.UUID
	Date   time.Time
}

type ListSlotsOutput struct {
	Slots []Slot
}

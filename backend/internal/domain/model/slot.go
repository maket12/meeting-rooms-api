package model

import (
	"fmt"
	"time"

	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/google/uuid"
)

var slotNameSpace = uuid.MustParse("01010101-0101-0101-0101-010101010101")

// ================ Rich model for Slot ================

type Slot struct {
	id     uuid.UUID
	roomID uuid.UUID
	start  time.Time
	end    time.Time
}

func NewSlot(
	roomID uuid.UUID,
	startTime time.Time,
) (*Slot, error) {
	if roomID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("room_id")
	}

	endTime := startTime.Add(30 * time.Minute)

	// generate unique uuid (depends on room_id and start_time)
	// which ensure id will be same every time for the same data
	data := fmt.Sprintf("%s|%d", roomID.String(), startTime.Unix())
	id := uuid.NewSHA1(slotNameSpace, []byte(data))

	return &Slot{
		id:     id,
		roomID: roomID,
		start:  startTime.UTC(),
		end:    endTime.UTC(),
	}, nil
}

func RestoreSlot(
	id, roomID uuid.UUID,
	startTime, endTime time.Time,
) *Slot {
	return &Slot{
		id:     id,
		roomID: roomID,
		start:  startTime,
		end:    endTime,
	}
}

// ================ Read-Only ================

func (s *Slot) ID() uuid.UUID     { return s.id }
func (s *Slot) RoomID() uuid.UUID { return s.roomID }
func (s *Slot) Start() time.Time  { return s.start }
func (s *Slot) End() time.Time    { return s.end }

package model

import (
	pkgerrs "backend/pkg/errs"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	dayVal         = 24 * time.Hour
	genSlotsWithin = dayVal * 7 // 7 days
	slotDuration   = 30         // amount of minutes
)

// ================ Custom time type ================

var (
	ErrInvalidHourValue    = errors.New("hour must be between 0 and 23")
	ErrInvalidMinutesValue = errors.New("minutes must be either 0 or 30")
)

type DayTime struct {
	totalMinutes int32
}

func NewDayTime(hmStr string) (DayTime, error) {
	parts := strings.Split(hmStr, ":")
	if len(parts) != 2 {
		return DayTime{}, pkgerrs.NewValueInvalidError("time_format")
	}

	h, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return DayTime{}, err
	}
	if h < 0 || h > 23 {
		return DayTime{}, pkgerrs.NewValueInvalidErrorWithReason(
			"time_format", ErrInvalidHourValue,
		)
	}

	m, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return DayTime{}, err
	}
	if m != 0 && m != 30 {
		return DayTime{}, pkgerrs.NewValueInvalidErrorWithReason(
			"time_format", ErrInvalidMinutesValue,
		)
	}

	return DayTime{totalMinutes: int32(h)*60 + int32(m)}, nil
}

func RestoreDayTime(minutes int32) DayTime {
	return DayTime{totalMinutes: minutes}
}

func (t DayTime) TotalMinutes() int32 { return t.totalMinutes }

func (t DayTime) String() string {
	return fmt.Sprintf("%02d:%02d", t.totalMinutes/60, t.totalMinutes%60)
}

func (t DayTime) Before(other DayTime) bool {
	return t.totalMinutes < other.totalMinutes
}

// ================ Rich model for Schedule ================

type Schedule struct {
	id         uuid.UUID
	roomID     uuid.UUID
	daysOfWeek []int
	startTime  DayTime
	endTime    DayTime
}

func NewSchedule(
	roomID uuid.UUID,
	daysOfWeek []int,
	startTime, endTime string,
) (*Schedule, error) {
	if roomID == uuid.Nil {
		return nil, pkgerrs.NewValueRequiredError("room_id")
	}
	if len(daysOfWeek) == 0 {
		return nil, pkgerrs.NewValueRequiredError("days_of_week")
	}

	dayMap := make(map[int]bool)
	for _, val := range daysOfWeek {
		if val < 1 || val > 7 {
			return nil, pkgerrs.NewValueInvalidError("days_of_week")
		}
		if dayMap[val] {
			return nil, pkgerrs.NewValueInvalidError("days_of_week")
		}
		dayMap[val] = true
	}

	startDateTime, err := NewDayTime(startTime)
	if err != nil {
		return nil, err
	}

	endDayTime, err := NewDayTime(endTime)
	if err != nil {
		return nil, err
	}

	if !startDateTime.Before(endDayTime) {
		return nil, pkgerrs.NewValueInvalidError("time_range")
	}

	return &Schedule{
		id:         uuid.New(),
		roomID:     roomID,
		daysOfWeek: daysOfWeek,
		startTime:  startDateTime,
		endTime:    endDayTime,
	}, nil
}

func RestoreSchedule(
	id, roomID uuid.UUID,
	daysOfWeek []int,
	startTime, endTime DayTime,
) *Schedule {
	return &Schedule{
		id:         id,
		roomID:     roomID,
		daysOfWeek: daysOfWeek,
		startTime:  startTime,
		endTime:    endTime,
	}
}

// ================ Read-Only ================

func (s *Schedule) ID() uuid.UUID      { return s.id }
func (s *Schedule) RoomID() uuid.UUID  { return s.roomID }
func (s *Schedule) DaysOfWeek() []int  { return s.daysOfWeek }
func (s *Schedule) StartTime() DayTime { return s.startTime }
func (s *Schedule) EndTime() DayTime   { return s.endTime }

// ================ Business logic ================

func (s *Schedule) createSlotsForDay(
	slotStorage []*Slot,
	day time.Time,
) ([]*Slot, error) {
	now := time.Now().UTC()

	start := s.startTime.TotalMinutes()
	end := s.endTime.TotalMinutes()

	for start+slotDuration <= end {
		slot, err := NewSlot(
			s.RoomID(),
			day.Add(time.Duration(start)*time.Minute),
		)
		if err != nil {
			return nil, err
		}

		// To ensure we look only at future slots
		if !slot.Start().Before(now) {
			slotStorage = append(slotStorage, slot)
		}

		start += slotDuration
	}

	return slotStorage, nil
}

func (s *Schedule) CreateSlots(date *time.Time) ([]*Slot, error) {
	var err error
	slots := make([]*Slot, 0)

	var dateFrom time.Time
	if date == nil {
		dateFrom = time.Now().UTC()
	} else {
		dateFrom = date.UTC()
	}

	// Round the time to not care about hours and minutes
	from := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	to := from.Add(genSlotsWithin)

	for day := from; day.Before(to); day = day.Add(dayVal) {
		weekDay := int(day.Weekday())
		if weekDay == 0 {
			weekDay = 7
		}

		// Check if the day is in schedule
		var contains bool
		for _, scheduleDay := range s.DaysOfWeek() {
			if weekDay == scheduleDay {
				contains = true
				break
			}
		}

		if contains {
			// If so, then create slots for the day
			slots, err = s.createSlotsForDay(slots, day)
			if err != nil {
				return nil, err
			}
		}
	}

	return slots, nil
}

package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDayTime(t *testing.T) {
	type testCase struct {
		name          string
		hmStr         string
		expectTotMins int32
		expectErr     error
	}

	var testCases = []testCase{
		{
			name:          "Success",
			hmStr:         "10:30",
			expectTotMins: 630,
			expectErr:     nil,
		},
		{
			name:      "Failure - invalid time format",
			hmStr:     "",
			expectErr: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "Failure - invalid hour value",
			hmStr:     "24:10",
			expectErr: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "Failure - invalid minutes value",
			hmStr:     "12:41",
			expectErr: pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dayTime, err := model.NewDayTime(tt.hmStr)
			if tt.expectErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.hmStr, dayTime.String())
				assert.Equal(t, tt.expectTotMins, dayTime.TotalMinutes())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}

func TestDayTime_Before(t *testing.T) {
	type testCase struct {
		first  model.DayTime
		second model.DayTime
		expect bool
	}

	var testCases = []testCase{
		{
			first:  model.RestoreDayTime(600),
			second: model.RestoreDayTime(720),
			expect: true,
		},
		{
			first:  model.RestoreDayTime(720),
			second: model.RestoreDayTime(600),
			expect: false,
		},
		{
			first:  model.RestoreDayTime(600),
			second: model.RestoreDayTime(630),
			expect: true,
		},
		{
			first:  model.RestoreDayTime(630),
			second: model.RestoreDayTime(600),
			expect: false,
		},
	}

	for n, tt := range testCases {
		testName := fmt.Sprintf("Test №%d", n)
		t.Run(testName, func(t *testing.T) {
			res := tt.first.Before(tt.second)
			require.Equal(t, tt.expect, res)
		})
	}
}

func TestNewSchedule(t *testing.T) {
	type testCase struct {
		name       string
		roomID     uuid.UUID
		daysOfWeek []int
		startTime  string
		endTime    string
		expect     error
	}

	var (
		testRoomID     = uuid.New()
		testDaysOfWeek = []int{1, 2, 3, 6}
		testStartTime  = "08:00"
		testEndTime    = "12:30"
	)

	var testCases = []testCase{
		{
			name:       "Success",
			roomID:     testRoomID,
			daysOfWeek: testDaysOfWeek,
			startTime:  testStartTime,
			endTime:    testEndTime,
			expect:     nil,
		},
		{
			name:   "Failure - null room id",
			roomID: uuid.Nil,
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:       "Failure - days of week are not specified",
			roomID:     testRoomID,
			daysOfWeek: make([]int, 0),
			expect:     pkgerrs.ErrValueIsRequired,
		},
		{
			name:       "Failure - days of week contain invalid value",
			roomID:     testRoomID,
			daysOfWeek: append(testDaysOfWeek, 10),
			expect:     pkgerrs.ErrValueIsInvalid,
		},
		{
			name:       "Failure - invalid time range",
			roomID:     testRoomID,
			daysOfWeek: testDaysOfWeek,
			startTime:  testEndTime,
			endTime:    testStartTime,
			expect:     pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := model.NewSchedule(
				tt.roomID,
				tt.daysOfWeek,
				tt.startTime,
				tt.endTime,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, schedule)

				assert.Equal(t, tt.roomID, schedule.RoomID())
				assert.ElementsMatch(t, tt.daysOfWeek, schedule.DaysOfWeek())
				assert.Equal(t, tt.startTime, schedule.StartTime().String())
				assert.Equal(t, tt.endTime, schedule.EndTime().String())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, schedule)
			}
		})
	}
}

func TestSchedule_CreateSlots_ExcludingCurrentDay(t *testing.T) {
	var (
		d           int
		days        []int
		currWeekDay = int(time.Now().UTC().Weekday())
	)

	if currWeekDay == 0 {
		currWeekDay = 7
	}

	for d = 1; d <= 7; d += 1 {
		if d != currWeekDay {
			days = append(days, d)
		}
	}

	schedule := model.RestoreSchedule(
		uuid.New(),
		uuid.New(),
		days,
		model.RestoreDayTime(690),
		model.RestoreDayTime(750),
	)

	slots, err := schedule.CreateSlots(nil)

	fstSlotTotalMins := slots[0].Start().Hour()*60 + slots[0].Start().Minute()
	sndSlotTotalMins := slots[len(slots)-1].End().Hour()*60 + slots[len(slots)-1].End().Minute()

	require.NoError(t, err)
	require.NotNil(t, slots)

	assert.Len(t, slots, 12)
	assert.Equal(
		t,
		int(schedule.StartTime().TotalMinutes()),
		fstSlotTotalMins,
	)
	assert.Equal(
		t,
		int(schedule.EndTime().TotalMinutes()),
		sndSlotTotalMins,
	)

	fmt.Printf("Slots[len=%d]:\n", len(slots))
	for i, slot := range slots {
		fmt.Printf("%d) start=%v end=%v\n",
			i+1, slot.Start(), slot.End())
	}
}

func TestSchedule_CreateSlots_CurrentDay(t *testing.T) {
	currWeekDay := int(time.Now().UTC().Weekday())
	if currWeekDay == 0 {
		currWeekDay = 7
	}

	schedule := model.RestoreSchedule(
		uuid.New(),
		uuid.New(),
		[]int{currWeekDay},
		model.RestoreDayTime(0),
		model.RestoreDayTime(1440),
	)

	slots, err := schedule.CreateSlots(nil)

	require.NoError(t, err)
	require.NotNil(t, slots)
	require.NotEmpty(t, slots)

	assert.True(t, slots[0].Start().Minute()%30 == 0)
	assert.True(t, slots[0].End().Minute()%30 == 0)
	assert.True(t, slots[len(slots)-1].Start().Minute()%30 == 0)
	assert.True(t, slots[len(slots)-1].End().Minute()%30 == 0)
}

func TestSchedule_CreateSlots_Error(t *testing.T) {
	schedule := model.RestoreSchedule(
		uuid.New(),
		uuid.Nil, // nullable room id
		[]int{1},
		model.RestoreDayTime(0),
		model.RestoreDayTime(1440),
	)

	slots, err := schedule.CreateSlots(nil)

	require.Error(t, err)
	require.Nil(t, slots)
}

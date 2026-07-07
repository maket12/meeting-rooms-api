package model_test

import (
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlot(t *testing.T) {
	type testCase struct {
		name      string
		roomID    uuid.UUID
		startTime time.Time
		expect    error
	}

	var (
		testRoomID    = uuid.New()
		testStartTime = time.Now()
	)

	var testCases = []testCase{
		{
			name:      "Success",
			roomID:    testRoomID,
			startTime: testStartTime,
			expect:    nil,
		},
		{
			name:   "Failure - invalid room id (nullable)",
			roomID: uuid.Nil,
			expect: pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			slot, err := model.NewSlot(tt.roomID, tt.startTime)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, slot)

				assert.False(t, slot.ID() == uuid.Nil)
				assert.Equal(t, tt.roomID, slot.RoomID())
				assert.Equal(t, tt.startTime.UTC(), slot.Start())
				assert.True(t, slot.End().Sub(slot.Start()) == 30*time.Minute)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, slot)
			}
		})
	}
}

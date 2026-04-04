package model_test

import (
	"MeetingRoomsAPI/internal/domain/model"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	pkgutils "MeetingRoomsAPI/pkg/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoom(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		roomName    string
		description *string
		capacity    *int
		expect      error
	}

	var (
		testRoomName    = "Lounge Zone №417"
		testDescription = pkgutils.VPtr("The very comfortable place to study or meet up in!")
		testCapacity    = pkgutils.VPtr(50)
	)

	var testCases = []testCase{
		{
			name:        "Success",
			roomName:    testRoomName,
			description: testDescription,
			capacity:    testCapacity,
			expect:      nil,
		},
		{
			name:     "Failure - room name is not specified",
			roomName: "",
			expect:   pkgerrs.ErrValueIsRequired,
		},
		{
			name:        "Failure - invalid description (empty string)",
			roomName:    testRoomName,
			description: pkgutils.VPtr(""),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "Failure - invalid capacity (negative number)",
			roomName:    testRoomName,
			description: testDescription,
			capacity:    pkgutils.VPtr(-100),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			room, err := model.NewRoom(
				tt.roomName,
				tt.description,
				tt.capacity,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, room)

				assert.False(t, room.ID() == uuid.Nil)
				assert.Equal(t, tt.roomName, room.Name())
				assert.Equal(t, tt.description, room.Description())
				assert.Equal(t, tt.capacity, room.Capacity())
				assert.False(t, room.CreatedAt().Before(time.Now().UTC()))
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, room)
			}
		})
	}
}

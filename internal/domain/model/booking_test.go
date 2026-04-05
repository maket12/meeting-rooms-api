package model_test

import (
	"MeetingRoomsAPI/internal/domain/model"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"MeetingRoomsAPI/pkg/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBooking(t *testing.T) {
	type testCase struct {
		name           string
		slotID         uuid.UUID
		userID         uuid.UUID
		conferenceLink *string
		expect         error
	}

	var (
		testSlotID         = uuid.New()
		testUserID         = uuid.New()
		testConferenceLink = utils.VPtr("https://zoom.com/ae45fgsaql")
	)

	var testCases = []testCase{
		{
			name:           "Success",
			slotID:         testSlotID,
			userID:         testUserID,
			conferenceLink: testConferenceLink,
			expect:         nil,
		},
		{
			name:   "Failure - nullable slot id",
			slotID: uuid.Nil,
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:   "Failure - nullable user id",
			slotID: testSlotID,
			userID: uuid.Nil,
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:           "Failure - invalid conference link(empty)",
			slotID:         testSlotID,
			userID:         testUserID,
			conferenceLink: utils.VPtr(""),
			expect:         pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			booking, err := model.NewBooking(
				tt.slotID,
				tt.userID,
				tt.conferenceLink,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, booking)

				assert.False(t, booking.ID() == uuid.Nil)
				assert.Equal(t, tt.slotID, booking.SlotID())
				assert.Equal(t, tt.userID, booking.UserID())
				assert.True(t, booking.Status() == model.BookingActive)
				assert.True(t, booking.Status().String() == string(model.BookingActive))
				assert.Same(t, tt.conferenceLink, booking.ConferenceLink())
				assert.NotEmpty(t, booking.CreatedAt())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, booking)
			}
		})
	}
}

func TestBooking_Cancel(t *testing.T) {
	type testCase struct {
		name        string
		booking     *model.Booking
		requestorID uuid.UUID
		expect      error
	}

	testUserID := uuid.New()

	var testCases = []testCase{
		{
			name: "Success",
			booking: model.RestoreBooking(
				uuid.New(),
				uuid.New(),
				testUserID,
				model.BookingActive,
				nil,
				time.Now().UTC(),
			),
			requestorID: testUserID,
			expect:      nil,
		},
		{
			name: "Success - already cancelled (idempotency)",
			booking: model.RestoreBooking(
				uuid.New(),
				uuid.New(),
				testUserID,
				model.BookingCancelled,
				nil,
				time.Now().UTC(),
			),
			requestorID: testUserID,
			expect:      nil,
		},
		{
			name: "Failure - forbidden",
			booking: model.RestoreBooking(
				uuid.New(),
				uuid.New(),
				testUserID,
				model.BookingActive,
				nil,
				time.Now().UTC(),
			),
			requestorID: uuid.New(),
			expect:      model.ErrBookingCantBeCancelled,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.booking.Cancel(tt.requestorID)
			if tt.expect == nil {
				require.NoError(t, err)
				require.True(t, tt.booking.Status() == model.BookingCancelled)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
			}
		})
	}
}

package usecase_test

import (
	"context"
	"errors"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListMyBookingsUC_Execute(t *testing.T) {
	type adapter struct {
		booking *mocks.MockBookingRepository
	}

	type testCase struct {
		name          string
		input         uuid.UUID
		mockBehaviour func(a adapter)
		expectErr     error
	}

	userID := uuid.New()
	dummyBookings := []*model.Booking{
		model.RestoreBooking(uuid.New(), uuid.New(), userID, "active", nil, time.Now().UTC()),
	}

	var tests = []testCase{
		{
			name:  "Success",
			input: userID,
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().ListByUserID(mock.Anything, userID).Return(dummyBookings, nil)
			},
			expectErr: nil,
		},
		{
			name:  "Failure - repository query error",
			input: userID,
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().ListByUserID(mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrListMyBookingsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			bookingRepo := mocks.NewMockBookingRepository(t)
			tt.mockBehaviour(adapter{booking: bookingRepo})

			// UC
			uc := usecase.NewListMyBookingsUC(bookingRepo)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, out.Bookings)
			}
		})
	}
}

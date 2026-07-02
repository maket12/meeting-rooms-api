package usecase_test

import (
	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/app/usecase"
	"backend/internal/domain/model"
	"backend/internal/domain/port/mocks"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelBookingUC_Execute(t *testing.T) {
	type adapter struct {
		booking *mocks.MockBookingRepository
	}

	type testCase struct {
		name          string
		input         dto.CancelBookingInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	bookingID := uuid.New()
	userID := uuid.New()

	dummyBooking := model.RestoreBooking(
		bookingID, uuid.New(), userID, "active",
		nil, time.Now().UTC(),
	)

	var tests = []testCase{
		{
			name: "Success",
			input: dto.CancelBookingInput{
				BookingID:   bookingID,
				RequestorID: userID,
			},
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().Get(mock.Anything, bookingID).Return(dummyBooking, nil)
				a.booking.EXPECT().UpdateStatus(mock.Anything, bookingID, mock.Anything).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - booking not found",
			input: dto.CancelBookingInput{
				BookingID:   bookingID,
				RequestorID: userID,
			},
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().Get(mock.Anything, bookingID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrBookingNotFound,
		},
		{
			name: "Failure - repository get error",
			input: dto.CancelBookingInput{
				BookingID:   bookingID,
				RequestorID: userID,
			},
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().Get(mock.Anything, bookingID).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetBookingDB,
		},
		{
			name: "Failure - repository update status error",
			input: dto.CancelBookingInput{
				BookingID:   bookingID,
				RequestorID: userID,
			},
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().Get(mock.Anything, bookingID).Return(dummyBooking, nil)
				a.booking.EXPECT().UpdateStatus(mock.Anything, bookingID, mock.Anything).Return(errors.New("db update error"))
			},
			expectErr: ucerrs.ErrUpdateBookingStatusDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			bookingRepo := mocks.NewMockBookingRepository(t)
			tt.mockBehaviour(adapter{booking: bookingRepo})

			// UC
			uc := usecase.NewCancelBookingUC(bookingRepo)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, out.Booking)
			}
		})
	}
}

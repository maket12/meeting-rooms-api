package usecase_test

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
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

func TestListBookingsUC_Execute(t *testing.T) {
	type adapter struct {
		booking *mocks.MockBookingRepository
	}

	type testCase struct {
		name          string
		input         dto.ListBookingsInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	dummyBookings := []*model.Booking{
		model.RestoreBooking(uuid.New(), uuid.New(), uuid.New(), "active", nil, time.Now().UTC()),
	}
	var expectedTotal int64 = 1

	var tests = []testCase{
		{
			name: "Success",
			input: dto.ListBookingsInput{
				Page:     1,
				PageSize: 10,
			},
			mockBehaviour: func(a adapter) {
				// Ожидаем вызов ListAll с лимитом 10 и офсетом 0 (так как (1-1)*10 = 0)
				a.booking.EXPECT().ListAll(mock.Anything, int32(10), int32(0)).
					Return(dummyBookings, expectedTotal, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - negative page",
			input: dto.ListBookingsInput{
				Page:     -1,
				PageSize: 10,
			},
			mockBehaviour: func(a adapter) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - negative page size",
			input: dto.ListBookingsInput{
				Page:     1,
				PageSize: -5,
			},
			mockBehaviour: func(a adapter) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - repository list error",
			input: dto.ListBookingsInput{
				Page:     2,
				PageSize: 20,
			},
			mockBehaviour: func(a adapter) {
				a.booking.EXPECT().ListAll(mock.Anything, int32(20), int32(20)).
					Return(nil, int64(0), errors.New("db scan error"))
			},
			expectErr: ucerrs.ErrListBookingsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			bookingRepo := mocks.NewMockBookingRepository(t)
			tt.mockBehaviour(adapter{booking: bookingRepo})

			// UC
			uc := usecase.NewListBookingsUC(bookingRepo)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int(expectedTotal), out.Pagination.Total)
				assert.NotEmpty(t, out.Bookings)
			}
		})
	}
}

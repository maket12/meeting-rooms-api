package usecase_test

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks" // Предположим, FakeTxManager лежит тут или в текущем пакете
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateBookingUC_Execute(t *testing.T) {
	type adapter struct {
		slot       *mocks.MockSlotRepository
		booking    *mocks.MockBookingRepository
		conference *mocks.MockConferenceService
	}

	type testCase struct {
		name          string
		input         dto.CreateBookingInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	slotID := uuid.New()
	userID := uuid.New()
	confLink := "https://conference.link/abc"

	futureSlot := model.RestoreSlot(slotID, uuid.New(), time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour))
	pastSlot := model.RestoreSlot(slotID, uuid.New(), time.Now().Add(-1*time.Hour), time.Now().Add(-30*time.Minute))
	dummyBooking := model.RestoreBooking(uuid.New(), slotID, userID, "active", nil, time.Now().UTC())

	var tests = []testCase{
		{
			name: "Success - with conference link",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: true,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(futureSlot, nil)
				a.conference.EXPECT().CreateMeeting(mock.Anything).Return(confLink, nil)
				a.booking.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Booking")).Return(dummyBooking, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - slot not found",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: false,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrSlotNotFound,
		},
		{
			name: "Failure - repository slot get error",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: false,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetSlotDB,
		},
		{
			name: "Failure - slot is in the past",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: false,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(pastSlot, nil)
			},
			expectErr: ucerrs.ErrCannotCreateBooking,
		},
		{
			name: "Failure - conference service error",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: true,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(futureSlot, nil)
				a.conference.EXPECT().CreateMeeting(mock.Anything).Return("", errors.New("rpc error"))
			},
			expectErr: ucerrs.ErrCreateMeeting,
		},
		{
			name: "Failure - booking already exists",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: false,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(futureSlot, nil)
				a.booking.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Booking")).Return(nil, pkgerrs.ErrObjectAlreadyExists)
			},
			expectErr: ucerrs.ErrBookingAlreadyExists,
		},
		{
			name: "Failure - repository booking create error",
			input: dto.CreateBookingInput{
				SlotID:               slotID,
				UserID:               userID,
				CreateConferenceLink: false,
			},
			mockBehaviour: func(a adapter) {
				a.slot.EXPECT().Get(mock.Anything, slotID).Return(futureSlot, nil)
				a.booking.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Booking")).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrCreateBookingDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализируем ваш FakeTxManager
			txManager := mocks.FakeTxManager{}

			slotRepo := mocks.NewMockSlotRepository(t)
			bookingRepo := mocks.NewMockBookingRepository(t)
			conferenceService := mocks.NewMockConferenceService(t)

			tt.mockBehaviour(adapter{
				slot:       slotRepo,
				booking:    bookingRepo,
				conference: conferenceService,
			})

			// Передаем txManager как первый аргумент
			uc := usecase.NewCreateBookingUC(txManager, slotRepo, bookingRepo, conferenceService)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, out)
			}
		})
	}
}

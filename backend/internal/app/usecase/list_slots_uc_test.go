package usecase_test

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSlotsUC_Execute(t *testing.T) {
	type adapter struct {
		room     *mocks.MockRoomRepository
		schedule *mocks.MockScheduleRepository
		slot     *mocks.MockSlotRepository
	}

	type testCase struct {
		name          string
		input         dto.ListSlotsInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	roomID := uuid.New()
	testDate := time.Now().Add(24 * time.Hour) // Завтра

	dummyRoom := model.RestoreRoom(roomID, "Conference Room", nil, nil, time.Now().UTC())
	dummySlots := []*model.Slot{
		model.RestoreSlot(uuid.New(), roomID, time.Now(), time.Now().Add(1*time.Hour)),
	}

	weekday := int(testDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	dummySchedule := model.RestoreSchedule(uuid.New(), roomID, []int{weekday}, model.RestoreDayTime(0), model.RestoreDayTime(180))

	var tests = []testCase{
		{
			name: "Success - slots already exist in DB",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(true, nil)
				a.slot.EXPECT().ListFree(mock.Anything, roomID, testDate).Return(dummySlots, nil)
			},
			expectErr: nil,
		},
		{
			name: "Success - no slots in DB, generated from schedule",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(false, nil)
				a.schedule.EXPECT().Get(mock.Anything, roomID).Return(dummySchedule, nil)
				a.slot.EXPECT().CreateBatch(mock.Anything, mock.AnythingOfType("[]*model.Slot")).Return(nil)
				a.slot.EXPECT().ListFree(mock.Anything, roomID, testDate).Return(dummySlots, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - room not found",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrRoomNotFound,
		},
		{
			name: "Failure - repository room get error",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(nil, errors.New("connection failed"))
			},
			expectErr: ucerrs.ErrGetRoomDB,
		},
		{
			name: "Failure - exists for date repository error",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(false, errors.New("db error"))
			},
			expectErr: ucerrs.ErrExistsForDateDB,
		},
		{
			name: "Failure - schedule not found when slots are empty",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(false, nil)
				a.schedule.EXPECT().Get(mock.Anything, roomID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrScheduleNotFound,
		},
		{
			name: "Failure - schedule repository get error",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(false, nil)
				a.schedule.EXPECT().Get(mock.Anything, roomID).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetScheduleDB,
		},
		{
			name: "Failure - batch creation of slots error",
			input: dto.ListSlotsInput{
				RoomID: roomID,
				Date:   testDate,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(dummyRoom, nil)
				a.slot.EXPECT().ExistsForDate(mock.Anything, roomID, testDate).Return(false, nil)
				a.schedule.EXPECT().Get(mock.Anything, roomID).Return(dummySchedule, nil)
				a.slot.EXPECT().CreateBatch(mock.Anything, mock.AnythingOfType("[]*model.Slot")).Return(errors.New("insert failed"))
			},
			expectErr: ucerrs.ErrCreateSlotsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := mocks.FakeTxManager{}

			roomRepo := mocks.NewMockRoomRepository(t)
			scheduleRepo := mocks.NewMockScheduleRepository(t)
			slotRepo := mocks.NewMockSlotRepository(t)

			tt.mockBehaviour(adapter{
				room:     roomRepo,
				schedule: scheduleRepo,
				slot:     slotRepo,
			})

			uc := usecase.NewListSlotsUC(txManager, roomRepo, scheduleRepo, slotRepo)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, out.Slots)
			}
		})
	}
}

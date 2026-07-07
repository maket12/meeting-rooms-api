package usecase_test

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	"testing"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type trmStub struct{}

func (s *trmStub) DoWithSettings(ctx context.Context, settings trm.Settings, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (s *trmStub) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func TestCreateScheduleUC_Execute(t *testing.T) {
	type adapter struct {
		room     *mocks.MockRoomRepository
		schedule *mocks.MockScheduleRepository
		slot     *mocks.MockSlotRepository
	}

	roomID := uuid.New()
	input := dto.CreateScheduleInput{
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	tests := []struct {
		name          string
		input         dto.CreateScheduleInput
		mockBehaviour func(a adapter)
		expectErr     error
	}{
		{
			name:  "Success",
			input: input,
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(&model.Room{}, nil)
				a.schedule.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Schedule")).
					RunAndReturn(func(ctx context.Context, s *model.Schedule) (*model.Schedule, error) {
						return s, nil
					})
				a.slot.EXPECT().CreateBatch(mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:  "Failure - Room Not Found",
			input: input,
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrRoomNotFound,
		},
		{
			name:  "Failure - Schedule Already Exists",
			input: input,
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Get(mock.Anything, roomID).Return(&model.Room{}, nil)
				a.schedule.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, pkgerrs.ErrObjectAlreadyExists)
			},
			expectErr: ucerrs.ErrScheduleAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				room:     mocks.NewMockRoomRepository(t),
				schedule: mocks.NewMockScheduleRepository(t),
				slot:     mocks.NewMockSlotRepository(t),
			}
			tt.mockBehaviour(a)

			// Передаем нашу заглушку trmStub{}
			uc := usecase.NewCreateScheduleUC(&trmStub{}, a.room, a.schedule, a.slot)
			_, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

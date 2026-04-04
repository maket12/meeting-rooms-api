package usecase_test

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/usecase"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/internal/domain/port/mocks"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateRoomUC_Execute(t *testing.T) {
	type adapter struct {
		room *mocks.MockRoomRepository
	}

	type testCase struct {
		name          string
		input         dto.CreateRoomInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.CreateRoomInput{
				Name:        "Room №147",
				Description: nil,
				Capacity:    nil,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Room")).
					RunAndReturn(func(ctx context.Context, r *model.Room) (*model.Room, error) {
						return r, nil
					})
			},
			expectErr: nil,
		},
		{
			name: "Failure - invalid input",
			input: dto.CreateRoomInput{
				Name:        "", // Empty string
				Description: nil,
				Capacity:    nil,
			},
			mockBehaviour: func(a adapter) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - repository error",
			input: dto.CreateRoomInput{
				Name:        "Room №017",
				Description: nil,
				Capacity:    nil,
			},
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Room")).
					Return(nil, errors.New("failed to create a room"))
			},
			expectErr: ucerrs.ErrCreateRoomDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			roomRepo := mocks.NewMockRoomRepository(t)
			tt.mockBehaviour(adapter{room: roomRepo})

			// UC
			uc := usecase.NewCreateRoomUC(roomRepo)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Name, out.Name)
			}
		})
	}
}

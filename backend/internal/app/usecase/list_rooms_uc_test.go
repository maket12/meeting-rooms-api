package usecase_test

import (
	"context"
	"errors"
	"testing"

	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRoomsUC_Execute(t *testing.T) {
	type adapter struct {
		room *mocks.MockRoomRepository
	}

	type testCase struct {
		name          string
		mockBehaviour func(a adapter)
		expectErr     error
	}

	var tests = []testCase{
		{
			name: "Success",
			mockBehaviour: func(a adapter) {
				room1, _ := model.NewRoom("Room №009", nil, nil)
				room2, _ := model.NewRoom("Room №007", nil, nil)
				mockRooms := []*model.Room{room1, room2}

				a.room.EXPECT().List(mock.Anything).Return(mockRooms, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - repository error",
			mockBehaviour: func(a adapter) {
				a.room.EXPECT().List(mock.Anything).
					Return(nil, errors.New("failed to get a list of rooms"))
			},
			expectErr: ucerrs.ErrListRoomsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			roomRepo := mocks.NewMockRoomRepository(t)
			tt.mockBehaviour(adapter{room: roomRepo})

			// UC
			uc := usecase.NewListRoomsUC(roomRepo)

			// Call method
			out, err := uc.Execute(context.Background())

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, out)
				assert.NotEmpty(t, out)
			}
		})
	}
}

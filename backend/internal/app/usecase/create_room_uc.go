package usecase

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
)

type CreateRoomUC struct {
	room port.RoomRepository
}

func NewCreateRoomUC(room port.RoomRepository) *CreateRoomUC {
	return &CreateRoomUC{room: room}
}

func (uc *CreateRoomUC) Execute(ctx context.Context, in dto.CreateRoomInput) (dto.CreateRoomOutput, error) {
	room, err := model.NewRoom(in.Name, in.Description, in.Capacity)
	if err != nil {
		return dto.CreateRoomOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	createdRoom, err := uc.room.Create(ctx, room)
	if err != nil {
		return dto.CreateRoomOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateRoomDB, err,
		)
	}

	return mapper.MapDomainToCreateRoomDTO(createdRoom), nil
}

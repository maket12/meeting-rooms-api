package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	"backend/internal/domain/port"
	"context"
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
		return dto.CreateRoomOutput{}, errs.Wrap(
			errs.ErrInvalidInput, err,
		)
	}

	createdRoom, err := uc.room.Create(ctx, room)
	if err != nil {
		return dto.CreateRoomOutput{}, errs.Wrap(
			errs.ErrCreateRoomDB, err,
		)
	}

	return mapper.MapDomainToCreateRoomDTO(createdRoom), nil
}

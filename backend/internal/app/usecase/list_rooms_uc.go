package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
	"context"
)

type ListRoomsUC struct {
	room port.RoomRepository
}

func NewListRoomsUC(room port.RoomRepository) *ListRoomsUC {
	return &ListRoomsUC{room: room}
}

func (uc *ListRoomsUC) Execute(ctx context.Context) (dto.ListRoomsOutput, error) {
	rooms, err := uc.room.List(ctx)
	if err != nil {
		return dto.ListRoomsOutput{}, errs.Wrap(errs.ErrListRoomsDB, err)
	}

	return mapper.MapDomainToListRoomsDTO(rooms), nil
}

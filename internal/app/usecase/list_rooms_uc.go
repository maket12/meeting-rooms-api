package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/port"
	"context"

	ucerrs "MeetingRoomsAPI/internal/app/errs"
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
		return dto.ListRoomsOutput{}, ucerrs.Wrap(ucerrs.ErrListRoomsDB, err)
	}

	return mapper.MapDomainToListRoomsDTO(rooms), nil
}

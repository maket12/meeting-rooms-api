package usecase

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
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

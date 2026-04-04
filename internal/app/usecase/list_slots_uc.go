package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/port"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type ListSlotsUC struct {
	trManager trm.Manager
	slot      port.SlotRepository
}

func NewListSlotsUC(
	trManager trm.Manager,
	slot port.SlotRepository,
) *ListSlotsUC {
	return &ListSlotsUC{
		trManager: trManager,
		slot:      slot,
	}
}

func (uc *ListSlotsUC) Execute(ctx context.Context, in dto.ListSlotsInput) (dto.ListSlotsOutput, error) {
	slots, err := uc.slot.ListFree(ctx, in.RoomID, in.Date)
	if err != nil {
		return dto.ListSlotsOutput{}, ucerrs.Wrap(
			ucerrs.ErrListSlotsDB, err,
		)
	}

	return mapper.MapDomainToListSlotsDTO(slots), nil
}

package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/port"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type ListSlotsUC struct {
	trManager trm.Manager
	room      port.RoomRepository
	schedule  port.ScheduleRepository
	slot      port.SlotRepository
}

func NewListSlotsUC(
	trManager trm.Manager,
	room port.RoomRepository,
	schedule port.ScheduleRepository,
	slot port.SlotRepository,
) *ListSlotsUC {
	return &ListSlotsUC{
		trManager: trManager,
		room:      room,
		schedule:  schedule,
		slot:      slot,
	}
}

func (uc *ListSlotsUC) Execute(ctx context.Context, in dto.ListSlotsInput) (dto.ListSlotsOutput, error) {
	_, err := uc.room.Get(ctx, in.RoomID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.ListSlotsOutput{}, ucerrs.ErrRoomNotFound
		}
		return dto.ListSlotsOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetRoomDB, err,
		)
	}

	slots, err := uc.slot.ListFree(ctx, in.RoomID, in.Date)
	if err != nil {
		return dto.ListSlotsOutput{}, ucerrs.Wrap(
			ucerrs.ErrListSlotsDB, err,
		)
	}

	if len(slots) == 0 {
		sch, err := uc.schedule.Get(ctx, in.RoomID)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) {
				return dto.ListSlotsOutput{}, ucerrs.ErrScheduleNotFound
			}
			return dto.ListSlotsOutput{}, ucerrs.Wrap(
				ucerrs.ErrGetScheduleDB, err,
			)
		}

		var worksToday bool

		w := int(in.Date.Weekday())
		if w == 0 {
			w = 7
		}

		for _, d := range sch.DaysOfWeek() {
			if d == w {
				worksToday = true
			}
		}

		if worksToday {
			slots, err = sch.CreateSlots()
			if err != nil {
				return dto.ListSlotsOutput{}, ucerrs.Wrap(
					ucerrs.ErrInvalidInput, err,
				)
			}

			if err := uc.slot.CreateBatch(ctx, slots); err != nil {
				return dto.ListSlotsOutput{}, ucerrs.Wrap(
					ucerrs.ErrCreateSlotsDB, err,
				)
			}
		}
	}

	return mapper.MapDomainToListSlotsDTO(slots), nil
}

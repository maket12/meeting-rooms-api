package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	port2 "backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type ListSlotsUC struct {
	trManager trm.Manager
	room      port2.RoomRepository
	schedule  port2.ScheduleRepository
	slot      port2.SlotRepository
}

func NewListSlotsUC(
	trManager trm.Manager,
	room port2.RoomRepository,
	schedule port2.ScheduleRepository,
	slot port2.SlotRepository,
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
			return dto.ListSlotsOutput{}, errs.ErrRoomNotFound
		}
		return dto.ListSlotsOutput{}, errs.Wrap(
			errs.ErrGetRoomDB, err,
		)
	}

	slots, err := uc.slot.ListFree(ctx, in.RoomID, in.Date)
	if err != nil {
		return dto.ListSlotsOutput{}, errs.Wrap(
			errs.ErrListSlotsDB, err,
		)
	}

	if len(slots) == 0 {
		sch, err := uc.schedule.Get(ctx, in.RoomID)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) {
				return dto.ListSlotsOutput{}, errs.ErrScheduleNotFound
			}
			return dto.ListSlotsOutput{}, errs.Wrap(
				errs.ErrGetScheduleDB, err,
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
				return dto.ListSlotsOutput{}, errs.Wrap(
					errs.ErrInvalidInput, err,
				)
			}

			if err := uc.slot.CreateBatch(ctx, slots); err != nil {
				return dto.ListSlotsOutput{}, errs.Wrap(
					errs.ErrCreateSlotsDB, err,
				)
			}
		}
	}

	return mapper.MapDomainToListSlotsDTO(slots), nil
}

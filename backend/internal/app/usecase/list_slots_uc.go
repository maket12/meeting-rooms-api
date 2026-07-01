package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
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
			return dto.ListSlotsOutput{}, errs.ErrRoomNotFound
		}
		return dto.ListSlotsOutput{}, errs.Wrap(
			errs.ErrGetRoomDB, err,
		)
	}

	var (
		slots              []*model.Slot
		listErr, createErr error
	)

	err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		slots, listErr = uc.slot.ListFree(ctx, in.RoomID, in.Date)
		if listErr != nil {
			return errs.Wrap(errs.ErrListSlotsDB, listErr)
		}

		if len(slots) == 0 {
			sch, getErr := uc.schedule.Get(ctx, in.RoomID)
			if getErr != nil {
				if errors.Is(getErr, pkgerrs.ErrObjectNotFound) {
					return errs.ErrScheduleNotFound
				}
				return errs.Wrap(errs.ErrGetScheduleDB, getErr)
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
				slots, createErr = sch.CreateSlots()
				if createErr != nil {
					return errs.Wrap(errs.ErrInvalidInput, createErr)
				}

				if createErr = uc.slot.CreateBatch(ctx, slots); createErr != nil {
					return errs.Wrap(errs.ErrCreateSlotsDB, createErr)
				}
			}
		}

		return nil
	})
	if err != nil {
		return dto.ListSlotsOutput{}, err
	}

	return mapper.MapDomainToListSlotsDTO(slots), nil
}

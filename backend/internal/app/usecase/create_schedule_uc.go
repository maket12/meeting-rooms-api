package usecase

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type CreateScheduleUC struct {
	trManager trm.Manager
	room      port.RoomRepository
	schedule  port.ScheduleRepository
	slot      port.SlotRepository
}

func NewCreateScheduleUC(
	trManager trm.Manager,
	room port.RoomRepository,
	schedule port.ScheduleRepository,
	slot port.SlotRepository,
) *CreateScheduleUC {
	return &CreateScheduleUC{
		trManager: trManager,
		room:      room,
		schedule:  schedule,
		slot:      slot,
	}
}

func (uc *CreateScheduleUC) Execute(ctx context.Context, in dto.CreateScheduleInput) (dto.CreateScheduleOutput, error) {
	var out dto.CreateScheduleOutput

	err := uc.trManager.Do(ctx, func(txCtx context.Context) error {
		_, getErr := uc.room.Get(txCtx, in.RoomID)
		if getErr != nil {
			if errors.Is(getErr, pkgerrs.ErrObjectNotFound) {
				return ucerrs.ErrRoomNotFound
			}
			return ucerrs.Wrap(ucerrs.ErrGetRoomDB, getErr)
		}

		// Create schedule
		schedule, createErr := model.NewSchedule(
			in.RoomID,
			in.DaysOfWeek,
			in.StartTime,
			in.EndTime,
		)
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, createErr)
		}

		createdSchedule, createErr := uc.schedule.Create(txCtx, schedule)
		if createErr != nil {
			if errors.Is(createErr, pkgerrs.ErrObjectAlreadyExists) {
				return ucerrs.ErrScheduleAlreadyExists
			}
			return ucerrs.Wrap(ucerrs.ErrCreateScheduleDB, createErr)
		}

		// Create slots
		slots, createErr := schedule.CreateSlots(nil)
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, createErr)
		}

		if createErr = uc.slot.CreateBatch(txCtx, slots); createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateSlotsDB, createErr)
		}

		out = mapper.MapDomainToCreateScheduleDTO(createdSchedule)

		return nil
	})

	if err != nil {
		return dto.CreateScheduleOutput{}, err
	}

	return out, nil
}

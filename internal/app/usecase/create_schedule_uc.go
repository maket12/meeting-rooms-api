package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/internal/domain/port"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"context"
	"errors"

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

	err := uc.trManager.Do(ctx, func(ctx context.Context) error {
		_, getErr := uc.room.Get(ctx, in.RoomID)
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

		createdSchedule, createErr := uc.schedule.Create(ctx, schedule)
		if createErr != nil {
			if errors.Is(createErr, pkgerrs.ErrObjectAlreadyExists) {
				return ucerrs.ErrScheduleAlreadyExists
			}
			return ucerrs.Wrap(ucerrs.ErrCreateScheduleDB, createErr)
		}

		// Create slots
		slots, createErr := schedule.CreateSlots()
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, createErr)
		}

		if createErr = uc.slot.CreateBatch(ctx, slots); createErr != nil {
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

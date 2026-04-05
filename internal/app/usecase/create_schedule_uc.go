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
		_, err := uc.room.Get(ctx, in.RoomID)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) {
				return ucerrs.ErrRoomNotFound
			}
			return ucerrs.Wrap(ucerrs.ErrGetRoomDB, err)
		}

		// Create schedule
		schedule, err := model.NewSchedule(
			in.RoomID,
			in.DaysOfWeek,
			in.StartTime,
			in.EndTime,
		)
		if err != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
		}

		createdSchedule, err := uc.schedule.Create(ctx, schedule)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
				return ucerrs.ErrScheduleAlreadyExists
			}
			return ucerrs.Wrap(ucerrs.ErrCreateScheduleDB, err)
		}

		// Create slots
		slots, err := schedule.CreateSlots()
		if err != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
		}

		if err := uc.slot.CreateBatch(ctx, slots); err != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateSlotsDB, err)
		}

		out = mapper.MapDomainToCreateScheduleDTO(createdSchedule)

		return nil
	})

	if err != nil {
		return dto.CreateScheduleOutput{}, err
	}

	return out, nil
}

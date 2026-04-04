package mapper

import (
	"MeetingRoomsAPI/internal/app/dto"
	"MeetingRoomsAPI/internal/domain/model"
)

func MapDomainToUserDTO(user *model.User) dto.User {
	return dto.User{
		ID:        user.ID(),
		Email:     user.Email(),
		Role:      user.Role().String(),
		CreatedAt: user.CreatedAt(),
	}
}

func MapDomainToRegisterDTO(user *model.User) dto.RegisterOutput {
	return dto.RegisterOutput{User: MapDomainToUserDTO(user)}
}

func MapDomainToRoomDTO(room *model.Room) dto.Room {
	return dto.Room{
		ID:          room.ID(),
		Name:        room.Name(),
		Description: room.Description(),
		Capacity:    room.Capacity(),
		CreatedAt:   room.CreatedAt(),
	}
}

func MapDomainToCreateRoomDTO(room *model.Room) dto.CreateRoomOutput {
	return dto.CreateRoomOutput(MapDomainToRoomDTO(room))
}

func MapDomainToListRoomsDTO(rooms []*model.Room) dto.ListRoomsOutput {
	mapped := make([]dto.Room, len(rooms))
	for i := range mapped {
		mapped[i] = MapDomainToRoomDTO(rooms[i])
	}

	return dto.ListRoomsOutput{Rooms: mapped}
}

func MapDomainToScheduleDTO(schedule *model.Schedule) dto.Schedule {
	return dto.Schedule{
		ID:         schedule.ID(),
		RoomID:     schedule.RoomID(),
		DaysOfWeek: schedule.DaysOfWeek(),
		StartTime:  schedule.StartTime().String(),
		EndTime:    schedule.EndTime().String(),
	}
}

func MapDomainToCreateScheduleDTO(schedule *model.Schedule) dto.CreateScheduleOutput {
	return dto.CreateScheduleOutput{
		Schedule: MapDomainToScheduleDTO(schedule),
	}
}

func MapDomainToSlotDTO(slot *model.Slot) dto.Slot {
	return dto.Slot{
		ID:     slot.ID(),
		RoomID: slot.RoomID(),
		Start:  slot.Start(),
		End:    slot.End(),
	}
}

func MapDomainToListSlotsDTO(slots []*model.Slot) dto.ListSlotsOutput {
	mapped := make([]dto.Slot, len(slots))
	for i := range mapped {
		mapped[i] = MapDomainToSlotDTO(slots[i])
	}

	return dto.ListSlotsOutput{Slots: mapped}
}

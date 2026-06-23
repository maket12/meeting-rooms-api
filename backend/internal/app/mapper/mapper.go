package mapper

import (
	dto2 "backend/internal/app/dto"
	model2 "backend/internal/domain/model"
)

func MapDomainToUserDTO(user *model2.User) dto2.User {
	return dto2.User{
		ID:        user.ID(),
		Email:     user.Email(),
		Role:      user.Role().String(),
		CreatedAt: user.CreatedAt(),
	}
}

func MapDomainToRegisterDTO(user *model2.User) dto2.RegisterOutput {
	return dto2.RegisterOutput{User: MapDomainToUserDTO(user)}
}

func MapDomainToRoomDTO(room *model2.Room) dto2.Room {
	return dto2.Room{
		ID:          room.ID(),
		Name:        room.Name(),
		Description: room.Description(),
		Capacity:    room.Capacity(),
		CreatedAt:   room.CreatedAt(),
	}
}

func MapDomainToCreateRoomDTO(room *model2.Room) dto2.CreateRoomOutput {
	return dto2.CreateRoomOutput{
		Room: MapDomainToRoomDTO(room),
	}
}

func MapDomainToListRoomsDTO(rooms []*model2.Room) dto2.ListRoomsOutput {
	mapped := make([]dto2.Room, len(rooms))
	for i := range mapped {
		mapped[i] = MapDomainToRoomDTO(rooms[i])
	}

	return dto2.ListRoomsOutput{Rooms: mapped}
}

func MapDomainToScheduleDTO(schedule *model2.Schedule) dto2.Schedule {
	return dto2.Schedule{
		ID:         schedule.ID(),
		RoomID:     schedule.RoomID(),
		DaysOfWeek: schedule.DaysOfWeek(),
		StartTime:  schedule.StartTime().String(),
		EndTime:    schedule.EndTime().String(),
	}
}

func MapDomainToCreateScheduleDTO(schedule *model2.Schedule) dto2.CreateScheduleOutput {
	return dto2.CreateScheduleOutput{
		Schedule: MapDomainToScheduleDTO(schedule),
	}
}

func MapDomainToSlotDTO(slot *model2.Slot) dto2.Slot {
	return dto2.Slot{
		ID:     slot.ID(),
		RoomID: slot.RoomID(),
		Start:  slot.Start(),
		End:    slot.End(),
	}
}

func MapDomainToListSlotsDTO(slots []*model2.Slot) dto2.ListSlotsOutput {
	mapped := make([]dto2.Slot, len(slots))
	for i := range mapped {
		mapped[i] = MapDomainToSlotDTO(slots[i])
	}

	return dto2.ListSlotsOutput{Slots: mapped}
}

func MapDomainToBookingDTO(booking *model2.Booking) dto2.Booking {
	return dto2.Booking{
		ID:             booking.ID(),
		SlotID:         booking.SlotID(),
		UserID:         booking.UserID(),
		Status:         booking.Status().String(),
		ConferenceLink: booking.ConferenceLink(),
		CreatedAt:      booking.CreatedAt(),
	}
}

func MapDomainToCreateBookingDTO(booking *model2.Booking) dto2.CreateBookingOutput {
	return dto2.CreateBookingOutput{Booking: MapDomainToBookingDTO(booking)}
}

func MapDomainToListBookings(bookings []*model2.Booking) dto2.ListBookingsOutput {
	mapped := make([]dto2.Booking, len(bookings))
	for i := range mapped {
		mapped[i] = MapDomainToBookingDTO(bookings[i])
	}
	return dto2.ListBookingsOutput{Bookings: mapped}
}

func MapDomainToListMyBookings(bookings []*model2.Booking) dto2.ListMyBookingsOutput {
	mapped := make([]dto2.Booking, len(bookings))
	for i := range mapped {
		mapped[i] = MapDomainToBookingDTO(bookings[i])
	}
	return dto2.ListMyBookingsOutput{Bookings: mapped}
}

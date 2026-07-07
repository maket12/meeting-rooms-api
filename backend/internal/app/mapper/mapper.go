package mapper

import (
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
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
	return dto.CreateRoomOutput{
		Room: MapDomainToRoomDTO(room),
	}
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

func MapDomainToBookingDTO(booking *model.Booking) dto.Booking {
	return dto.Booking{
		ID:             booking.ID(),
		SlotID:         booking.SlotID(),
		UserID:         booking.UserID(),
		Status:         booking.Status().String(),
		ConferenceLink: booking.ConferenceLink(),
		CreatedAt:      booking.CreatedAt(),
	}
}

func MapDomainToCreateBookingDTO(booking *model.Booking) dto.CreateBookingOutput {
	return dto.CreateBookingOutput{Booking: MapDomainToBookingDTO(booking)}
}

func MapDomainToListBookings(bookings []*model.Booking) dto.ListBookingsOutput {
	mapped := make([]dto.Booking, len(bookings))
	for i := range mapped {
		mapped[i] = MapDomainToBookingDTO(bookings[i])
	}
	return dto.ListBookingsOutput{Bookings: mapped}
}

func MapDomainToListMyBookings(bookings []*model.Booking) dto.ListMyBookingsOutput {
	mapped := make([]dto.Booking, len(bookings))
	for i := range mapped {
		mapped[i] = MapDomainToBookingDTO(bookings[i])
	}
	return dto.ListMyBookingsOutput{Bookings: mapped}
}

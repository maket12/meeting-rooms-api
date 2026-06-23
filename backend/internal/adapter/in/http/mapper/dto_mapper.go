package mapper

import (
	httpdto "backend/internal/adapter/in/http/dto"
	appdto "backend/internal/app/dto"
	"time"

	"github.com/google/uuid"
)

func MapRequestToDummyLogin(request httpdto.DummyLoginRequest) appdto.DummyLoginInput {
	return appdto.DummyLoginInput{Role: request.Role}
}

func MapDummyLoginToResponse(output appdto.DummyLoginOutput) httpdto.DummyLoginResponse {
	return httpdto.DummyLoginResponse{Token: output.Token}
}

func MapRequestToRegister(request httpdto.RegisterRequest) appdto.RegisterInput {
	return appdto.RegisterInput{
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
	}
}

func MapUserToResponse(user appdto.User) httpdto.UserResponse {
	return httpdto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}

func MapRegisterToResponse(output appdto.RegisterOutput) httpdto.RegisterResponse {
	return httpdto.RegisterResponse{User: MapUserToResponse(output.User)}
}

func MapRequestToLogin(request httpdto.LoginRequest) appdto.LoginInput {
	return appdto.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	}
}

func MapLoginToResponse(output appdto.LoginOutput) httpdto.LoginResponse {
	return httpdto.LoginResponse{Token: output.Token}
}

func MapRoomToResponse(room appdto.Room) httpdto.RoomResponse {
	return httpdto.RoomResponse{
		ID:          room.ID.String(),
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
		CreatedAt:   room.CreatedAt.String(),
	}
}

func MapRequestToCreateRoom(request httpdto.CreateRoomRequest) appdto.CreateRoomInput {
	return appdto.CreateRoomInput{
		Name:        request.Name,
		Description: request.Description,
		Capacity:    request.Capacity,
	}
}

func MapCreateRoomToResponse(output appdto.CreateRoomOutput) httpdto.CreateRoomResponse {
	return httpdto.CreateRoomResponse{
		Room: MapRoomToResponse(output.Room),
	}
}

func MapListRoomsToResponse(output appdto.ListRoomsOutput) httpdto.ListRoomsResponse {
	rooms := make([]httpdto.RoomResponse, len(output.Rooms))
	for i := range rooms {
		rooms[i] = MapRoomToResponse(output.Rooms[i])
	}

	return httpdto.ListRoomsResponse{Rooms: rooms}
}

func MapScheduleToResponse(schedule appdto.Schedule) httpdto.ScheduleResponse {
	return httpdto.ScheduleResponse{
		ID:         schedule.ID.String(),
		RoomID:     schedule.RoomID.String(),
		DaysOfWeek: schedule.DaysOfWeek,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}
}

func MapRequestToCreateSchedule(request httpdto.CreateScheduleRequest) appdto.CreateScheduleInput {
	roomID, _ := uuid.Parse(request.RoomID)
	return appdto.CreateScheduleInput{
		RoomID:     roomID,
		DaysOfWeek: request.DaysOfWeek,
		StartTime:  request.StartTime,
		EndTime:    request.EndTime,
	}
}

func MapCreateScheduleToResponse(output appdto.CreateScheduleOutput) httpdto.CreateScheduleResponse {
	return httpdto.CreateScheduleResponse{
		Schedule: MapScheduleToResponse(output.Schedule),
	}
}

func MapRequestToListSlots(request httpdto.ListSlotsRequest) (appdto.ListSlotsInput, error) {
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		return appdto.ListSlotsInput{}, err
	}

	return appdto.ListSlotsInput{
		RoomID: uuid.MustParse(request.RoomID),
		Date:   date,
	}, nil
}

func MapSlotToResponse(slot appdto.Slot) httpdto.SlotResponse {
	return httpdto.SlotResponse{
		ID:     slot.ID.String(),
		RoomID: slot.RoomID.String(),
		Start:  slot.Start.String(),
		End:    slot.End.String(),
	}
}

func MapListSlotsToResponse(output appdto.ListSlotsOutput) httpdto.ListSlotsResponse {
	slots := make([]httpdto.SlotResponse, len(output.Slots))
	for i := range slots {
		slots[i] = MapSlotToResponse(output.Slots[i])
	}

	return httpdto.ListSlotsResponse{Slots: slots}
}

func MapBookingToResponse(booking appdto.Booking) httpdto.BookingResponse {
	return httpdto.BookingResponse{
		ID:             booking.ID.String(),
		SlotID:         booking.SlotID.String(),
		UserID:         booking.UserID.String(),
		Status:         booking.Status,
		ConferenceLink: booking.ConferenceLink,
		CreatedAt:      booking.CreatedAt.Format(time.RFC3339),
	}
}

func MapRequestToCreateBooking(request httpdto.CreateBookingRequest, userID uuid.UUID) appdto.CreateBookingInput {
	slotID, _ := uuid.Parse(request.SlotID)
	return appdto.CreateBookingInput{
		SlotID:               slotID,
		UserID:               userID,
		CreateConferenceLink: request.CreateConferenceLink,
	}
}

func MapCreateBookingToResponse(output appdto.CreateBookingOutput) httpdto.CreateBookingResponse {
	return httpdto.CreateBookingResponse{
		Booking: MapBookingToResponse(output.Booking),
	}
}

func MapCancelBookingToResponse(output appdto.CancelBookingOutput) httpdto.CancelBookingResponse {
	return httpdto.CancelBookingResponse{
		Booking: MapBookingToResponse(output.Booking),
	}
}

func MapBookingsToResponse(bookings []appdto.Booking) []httpdto.BookingResponse {
	response := make([]httpdto.BookingResponse, len(bookings))
	for i := range bookings {
		response[i] = MapBookingToResponse(bookings[i])
	}
	return response
}

func MapListMyBookingsToResponse(output appdto.ListMyBookingsOutput) httpdto.ListMyBookingsResponse {
	return httpdto.ListMyBookingsResponse{
		Bookings: MapBookingsToResponse(output.Bookings),
	}
}

func MapPaginationToResponse(pagination appdto.Pagination) httpdto.PaginationResponse {
	return httpdto.PaginationResponse{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    pagination.Total,
	}
}

func MapListToResponse(output appdto.ListBookingsOutput) httpdto.ListBookingsResponse {
	return httpdto.ListBookingsResponse{
		Bookings:   MapBookingsToResponse(output.Bookings),
		Pagination: MapPaginationToResponse(output.Pagination),
	}
}

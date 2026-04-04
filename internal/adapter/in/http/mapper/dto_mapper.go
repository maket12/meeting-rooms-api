package mapper

import (
	httpdto "MeetingRoomsAPI/internal/adapter/in/http/dto"
	appdto "MeetingRoomsAPI/internal/app/dto"
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
	room := appdto.Room(output)
	return httpdto.CreateRoomResponse(MapRoomToResponse(room))
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
	return appdto.CreateScheduleInput{
		RoomID:     uuid.MustParse(request.RoomID),
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

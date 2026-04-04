package dto

type DummyLoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RoomResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
	CreatedAt   string  `json:"created_at"`
}

type CreateRoomResponse RoomResponse

type ListRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type ScheduleResponse struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}

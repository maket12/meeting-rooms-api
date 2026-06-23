package dto

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type CreateScheduleRequest struct {
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ListSlotsRequest struct {
	RoomID string `json:"roomId"`
	Date   string `json:"date"`
}

type CreateBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type CancelBookingRequest struct {
	BookingID string `json:"bookingId"`
}

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
	RoomID     string `param:"room_id"`
	DaysOfWeek []int  `json:"days_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

type ListSlotsRequest struct {
	RoomID string `param:"room_id"`
	Date   string `query:"date"`
}

type CreateBookingRequest struct {
	SlotID               string `json:"slot_id"`
	CreateConferenceLink bool   `json:"create_conference_link"`
}

type CancelBookingRequest struct {
	BookingID string `param:"booking_id"`
}

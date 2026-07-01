package dto

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
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

type CreateRoomResponse struct {
	Room RoomResponse `json:"room"`
}

type ListRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type ScheduleResponse struct {
	ID         string `json:"id"`
	RoomID     string `json:"room_id"`
	DaysOfWeek []int  `json:"days_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}

type SlotResponse struct {
	ID     string `json:"id"`
	RoomID string `json:"room_id"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

type ListSlotsResponse struct {
	Slots []SlotResponse `json:"slots"`
}

type BookingResponse struct {
	ID             string  `json:"id"`
	SlotID         string  `json:"slot_id"`
	UserID         string  `json:"user_id"`
	Status         string  `json:"status"`
	ConferenceLink *string `json:"conference_link"`
	CreatedAt      string  `json:"created_at"`
}

type PaginationResponse struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type CreateBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}

type CancelBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}

type ListBookingsResponse struct {
	Bookings   []BookingResponse  `json:"bookings"`
	Pagination PaginationResponse `json:"pagination"`
}

type ListMyBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
}

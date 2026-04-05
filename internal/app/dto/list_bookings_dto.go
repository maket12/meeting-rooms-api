package dto

type ListBookingsInput struct {
	Page     int
	PageSize int
}

type ListBookingsOutput struct {
	Bookings   []Booking
	Pagination Pagination
}

package dto

import "github.com/google/uuid"

type CancelBookingInput struct {
	BookingID   uuid.UUID
	RequestorID uuid.UUID
}

type CancelBookingOutput struct {
	Booking Booking
}

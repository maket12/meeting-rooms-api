package dto

import "github.com/google/uuid"

type CreateBookingInput struct {
	SlotID               uuid.UUID
	UserID               uuid.UUID
	CreateConferenceLink bool
}

type CreateBookingOutput struct {
	Booking Booking
}

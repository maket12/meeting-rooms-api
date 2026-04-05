package model

import (
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrBookingCantBeCancelled = errors.New("booking cannot be cancelled")

type BookingStatus string

const (
	BookingActive    BookingStatus = "active"
	BookingCancelled BookingStatus = "cancelled"
)

func (s BookingStatus) String() string { return string(s) }

// ================ Rich model for Booking ================

type Booking struct {
	id             uuid.UUID
	slotID         uuid.UUID
	userID         uuid.UUID
	status         BookingStatus
	conferenceLink *string
	createdAt      time.Time
}

func NewBooking(
	slotID, userID uuid.UUID,
	conferenceLink *string,
) (*Booking, error) {
	if slotID == uuid.Nil {
		return nil, pkgerrs.NewValueRequiredError("slot_id")
	}
	if userID == uuid.Nil {
		return nil, pkgerrs.NewValueRequiredError("user_id")
	}

	if conferenceLink != nil && *conferenceLink == "" {
		return nil, pkgerrs.NewValueInvalidError("conference_link")
	}

	return &Booking{
		id:             uuid.New(),
		slotID:         slotID,
		userID:         userID,
		status:         BookingActive,
		conferenceLink: conferenceLink,
		createdAt:      time.Now().UTC(),
	}, nil
}

func RestoreBooking(
	id, slotID, userID uuid.UUID,
	status BookingStatus,
	conferenceLink *string,
	createdAt time.Time,
) *Booking {
	return &Booking{
		id:             id,
		slotID:         slotID,
		userID:         userID,
		status:         status,
		conferenceLink: conferenceLink,
		createdAt:      createdAt,
	}
}

// ================ Read-Only ================

func (b *Booking) ID() uuid.UUID           { return b.id }
func (b *Booking) SlotID() uuid.UUID       { return b.slotID }
func (b *Booking) UserID() uuid.UUID       { return b.userID }
func (b *Booking) Status() BookingStatus   { return b.status }
func (b *Booking) ConferenceLink() *string { return b.conferenceLink }
func (b *Booking) CreatedAt() time.Time    { return b.createdAt }

// ================ Business logic ================

func (b *Booking) Cancel(requestorID uuid.UUID) error {
	if b.userID != requestorID {
		return ErrBookingCantBeCancelled
	}

	if b.Status() == BookingCancelled {
		return nil
	}

	b.status = BookingCancelled
	return nil
}

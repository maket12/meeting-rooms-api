package model

import (
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"time"

	"github.com/google/uuid"
)

// ================ Rich model for Room ================

type Room struct {
	id          uuid.UUID
	name        string
	description *string
	capacity    *int
	createdAt   time.Time
}

func NewRoom(
	name string,
	description *string,
	capacity *int,
) (*Room, error) {
	if name == "" {
		return nil, pkgerrs.NewValueRequiredError("testName")
	}
	if description != nil && *description == "" {
		return nil, pkgerrs.NewValueInvalidError("description")
	}
	if capacity != nil && *capacity < 0 {
		return nil, pkgerrs.NewValueInvalidError("capacity")
	}

	return &Room{
		id:          uuid.New(),
		name:        name,
		description: description,
		capacity:    capacity,
		createdAt:   time.Now().UTC(),
	}, nil
}

func RestoreRoom(
	id uuid.UUID,
	name string,
	description *string,
	capacity *int,
	createdAt time.Time,
) *Room {
	return &Room{
		id:          id,
		name:        name,
		description: description,
		capacity:    capacity,
		createdAt:   createdAt,
	}
}

// ================ Read-Only ================

func (r *Room) ID() uuid.UUID        { return r.id }
func (r *Room) Name() string         { return r.name }
func (r *Room) Description() *string { return r.description }
func (r *Room) Capacity() *int       { return r.capacity }
func (r *Room) CreatedAt() time.Time { return r.createdAt }

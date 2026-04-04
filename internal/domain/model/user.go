package model

import (
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

func (r UserRole) String() string { return string(r) }

// ================ Rich model for User ================

type User struct {
	id           uuid.UUID
	email        string
	passwordHash string
	role         UserRole
	createdAt    time.Time
}

func NewUser(email, passwordHash string, role UserRole) (*User, error) {
	if email == "" {
		return nil, pkgerrs.NewValueRequiredError("email")
	}
	if passwordHash == "" {
		return nil, pkgerrs.NewValueRequiredError("password_hash")
	}

	switch {
	case role == "":
		return nil, pkgerrs.NewValueRequiredError("role")
	case role != RoleAdmin && role != RoleUser:
		return nil, pkgerrs.NewValueInvalidError("role")
	}

	return &User{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    time.Now().UTC(),
	}, nil
}

func RestoreUser(
	id uuid.UUID,
	email, passwordHash string,
	role UserRole,
	createdAt time.Time,
) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    createdAt,
	}
}

// ================ Read-Only ================

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Email() string        { return u.email }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) Role() UserRole       { return u.role }
func (u *User) CreatedAt() time.Time { return u.createdAt }

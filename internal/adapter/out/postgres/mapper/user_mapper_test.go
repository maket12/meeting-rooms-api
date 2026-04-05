package mapper_test

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/mapper"
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserToSQLCCreate(t *testing.T) {
	user, _ := model.NewUser(
		"amazing-email@avito.ru",
		"new-pass",
		model.RoleUser,
	)

	mapped := mapper.MapUserToSQLCCreate(user)

	require.True(t, mapped.ID.Valid)
	require.True(t, mapped.CreatedAt.Valid)

	assert.Equal(t, [16]byte(user.ID()), mapped.ID.Bytes)
	assert.Equal(t, user.Email(), mapped.Email)
	assert.Equal(t, user.PasswordHash(), mapped.PasswordHash)
	assert.Equal(t, user.Role().String(), mapped.Role)
	assert.Equal(t, user.CreatedAt(), mapped.CreatedAt.Time)
}

func TestMapSQLCToUser(t *testing.T) {
	rawUser := sqlc.User{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Email:        "amazing-email@avito.ru",
		PasswordHash: "hashed_password",
		Role:         "admin",
		CreatedAt: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}

	user := mapper.MapSQLCToUser(rawUser)

	require.NotNil(t, user)

	assert.Equal(t, rawUser.ID.Bytes, [16]byte(user.ID()))
	assert.Equal(t, rawUser.Email, user.Email())
	assert.Equal(t, rawUser.PasswordHash, user.PasswordHash())
	assert.Equal(t, rawUser.Role, user.Role().String())
	assert.Equal(t, rawUser.CreatedAt.Time, user.CreatedAt())
}

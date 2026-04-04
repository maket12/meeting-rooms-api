package mapper

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapUserToSQLCCreate(user *model.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: user.ID(),
			Valid: true,
		},
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
		Role:         user.Role().String(),
		CreatedAt: pgtype.Timestamptz{
			Time:             user.CreatedAt(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}
}

func MapSQLCToUser(raw sqlc.User) *model.User {
	return model.RestoreUser(
		raw.ID.Bytes,
		raw.Email,
		raw.PasswordHash,
		model.UserRole(raw.Role),
		raw.CreatedAt.Time.UTC(),
	)
}

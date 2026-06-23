package mapper

import (
	sqlc2 "backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/model"
	"backend/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapRoomToSQLCCreate(room *model.Room) sqlc2.CreateRoomParams {
	var (
		desc     pgtype.Text
		capacity int32
	)
	if room.Description() != nil {
		desc = pgtype.Text{
			String: *room.Description(),
			Valid:  true,
		}
	}
	if room.Capacity() != nil {
		capacity = int32(*room.Capacity())
	}

	return sqlc2.CreateRoomParams{
		ID: pgtype.UUID{
			Bytes: room.ID(),
			Valid: true,
		},
		Name:        room.Name(),
		Description: desc,
		Capacity:    capacity,
		CreatedAt: pgtype.Timestamptz{
			Time:             room.CreatedAt(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}
}

func MapSQLCToRoom(rawRoom sqlc2.Room) *model.Room {
	var desc *string
	if rawRoom.Description.Valid {
		desc = &rawRoom.Description.String
	}

	return model.RestoreRoom(
		rawRoom.ID.Bytes,
		rawRoom.Name,
		desc,
		utils.VPtr(int(rawRoom.Capacity)),
		rawRoom.CreatedAt.Time.UTC(),
	)
}

func MapSQLCToRoomsList(rawRooms []sqlc2.Room) []*model.Room {
	rooms := make([]*model.Room, len(rawRooms))
	for i := range rooms {
		mapped := MapSQLCToRoom(rawRooms[i])
		rooms[i] = mapped
	}
	return rooms
}

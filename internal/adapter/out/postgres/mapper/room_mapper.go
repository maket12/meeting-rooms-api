package mapper

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapRoomToSQLCCreate(room *model.Room) sqlc.CreateRoomParams {
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

	return sqlc.CreateRoomParams{
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

func MapSQLCToRoom(rawRoom sqlc.Room) *model.Room {
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

func MapSQLCToRoomsList(rawRooms []sqlc.Room) []*model.Room {
	rooms := make([]*model.Room, 0, len(rawRooms))
	for _, rawRoom := range rawRooms {
		mappedRoom := MapSQLCToRoom(rawRoom)
		rooms = append(rooms, mappedRoom)
	}
	return rooms
}

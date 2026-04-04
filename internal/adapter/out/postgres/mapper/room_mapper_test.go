package mapper_test

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/mapper"
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/pkg/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapRoomToSQLCCreate(t *testing.T) {
	t.Parallel()

	room, _ := model.NewRoom(
		"Room №001",
		utils.VPtr("The fancy and vibe area"),
		utils.VPtr(100),
	)

	mapped := mapper.MapRoomToSQLCCreate(room)

	require.True(t, mapped.Description.Valid)
	assert.Equal(t, [16]byte(room.ID()), mapped.ID.Bytes)
	assert.Equal(t, room.Name(), mapped.Name)
	assert.Equal(t, *room.Description(), mapped.Description.String)
	assert.Equal(t, int32(*room.Capacity()), mapped.Capacity)
	assert.Equal(t, room.CreatedAt(), mapped.CreatedAt.Time)
}

func TestMapSQLCToRoom(t *testing.T) {
	t.Parallel()

	rawRoom := sqlc.Room{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Name: "Coworking area №2",
		Description: pgtype.Text{
			String: "An open space to study or work",
			Valid:  true,
		},
		Capacity: 500,
		CreatedAt: pgtype.Timestamptz{
			Time:             time.Now().UTC(),
			InfinityModifier: 0,
			Valid:            true,
		},
	}

	room := mapper.MapSQLCToRoom(rawRoom)

	require.NotNil(t, room)
	require.NotNil(t, room.Description())
	require.NotNil(t, room.Capacity())

	assert.Equal(t, rawRoom.ID.Bytes, [16]byte(room.ID()))
	assert.Equal(t, rawRoom.Name, room.Name())
	assert.Equal(t, rawRoom.Description.String, *room.Description())
	assert.Equal(t, rawRoom.Capacity, int32(*room.Capacity()))
	assert.Equal(t, rawRoom.CreatedAt.Time, room.CreatedAt())
}

func TestMapSQLCToRoomsList(t *testing.T) {
	t.Parallel()

	rawRooms := []sqlc.Room{
		{
			ID: pgtype.UUID{
				Bytes: uuid.New(),
				Valid: true,
			},
			Name: "Coworking area №2",
			Description: pgtype.Text{
				String: "An open space to study or work",
				Valid:  true,
			},
			Capacity: 500,
			CreatedAt: pgtype.Timestamptz{
				Time:             time.Now().UTC(),
				InfinityModifier: 0,
				Valid:            true,
			},
		},
		{
			ID: pgtype.UUID{
				Bytes: uuid.New(),
				Valid: true,
			},
			Name: "Coworking area №3",
			Description: pgtype.Text{
				String: "An open space to study or work",
				Valid:  true,
			},
			Capacity: 500,
			CreatedAt: pgtype.Timestamptz{
				Time:             time.Now().UTC(),
				InfinityModifier: 0,
				Valid:            true,
			},
		},
	}

	mapped := mapper.MapSQLCToRoomsList(rawRooms)

	require.NotNil(t, mapped)
	require.NotEmpty(t, mapped)
	require.Len(t, mapped, len(rawRooms))

	for i := 0; i < len(rawRooms); i++ {
		require.NotNil(t, mapped[i])
		assert.Equal(t, rawRooms[i].ID.Bytes, [16]byte(mapped[i].ID()))
		assert.Equal(t, rawRooms[i].Name, mapped[i].Name())
		// ...
		assert.Equal(t, rawRooms[i].CreatedAt.Time, mapped[i].CreatedAt())
	}
}

package postgres

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/mapper"
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	pkgpostgres "MeetingRoomsAPI/pkg/postgres"
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	q      *sqlc.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewRoomRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *RoomRepository {
	return &RoomRepository{
		q:      sqlc.New(),
		pool:   pgClient.Pool,
		getter: getter,
	}
}

func (r *RoomRepository) Create(ctx context.Context, room *model.Room) (*model.Room, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	params := mapper.MapRoomToSQLCCreate(room)

	rawRoom, err := r.q.CreateRoom(ctx, db, params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToRoom(rawRoom), nil
}

func (r *RoomRepository) Get(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawRoom, err := r.q.GetRoomByID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("room", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToRoom(rawRoom), nil
}

func (r *RoomRepository) List(ctx context.Context) ([]*model.Room, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawRooms, err := r.q.ListRooms(ctx, db)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToRoomsList(rawRooms), nil
}

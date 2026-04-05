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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	q      *sqlc.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewScheduleRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *ScheduleRepository {
	return &ScheduleRepository{
		q:      sqlc.New(),
		pool:   pgClient.Pool,
		getter: getter,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	params := mapper.MapScheduleToSQLCCreate(schedule)

	rawSchedule, err := r.q.CreateSchedule(ctx, db, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, pkgerrs.NewObjectAlreadyExistsErrorWithReason(
					"schedule", pgErr,
				)
			}
		}
		return nil, err
	}

	return mapper.MapSQLCToSchedule(rawSchedule), nil
}

func (r *ScheduleRepository) Get(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	rawSchedule, err := r.q.GetScheduleByRoomID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: roomID,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("schedule", roomID)
		}
		return nil, err
	}

	return mapper.MapSQLCToSchedule(rawSchedule), nil
}

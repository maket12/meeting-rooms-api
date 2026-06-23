package postgres

import (
	"backend/internal/adapter/out/postgres/mapper"
	sqlc2 "backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/model"
	pkgerrs "backend/pkg/errs"
	pkgpostgres "backend/pkg/postgres"
	"context"
	"errors"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct {
	q      *sqlc2.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewSlotRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *SlotRepository {
	return &SlotRepository{
		q:      sqlc2.New(),
		pool:   pgClient.Pool,
		getter: getter,
	}
}

func (r *SlotRepository) CreateBatch(ctx context.Context, slots []*model.Slot) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	params := mapper.MapSlotsToSQLCCreateBatch(slots)
	if _, err := r.q.CreateSlotsBatch(ctx, db, params); err != nil {
		return err
	}
	return nil
}

func (r *SlotRepository) Get(ctx context.Context, id uuid.UUID) (*model.Slot, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	rawSlot, err := r.q.GetSlotByID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("slot", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToSlot(rawSlot), nil
}

func (r *SlotRepository) ListFree(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*model.Slot, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	rawSlots, err := r.q.GetFreeSlotsByRoomAndDate(
		ctx,
		db,
		sqlc2.GetFreeSlotsByRoomAndDateParams{
			RoomID: pgtype.UUID{
				Bytes: roomID,
				Valid: true,
			},
			Date: pgtype.Date{
				Time:             date,
				InfinityModifier: 0,
				Valid:            true,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToSlots(rawSlots), nil
}

package postgres

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/adapter/out/postgres/mapper"
	"github.com/maket12/meeting-rooms-api/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	pkgpostgres "github.com/maket12/meeting-rooms-api/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	q      *sqlc.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewBookingRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *BookingRepository {
	return &BookingRepository{
		q:      sqlc.New(),
		pool:   pgClient.Pool,
		getter: getter,
	}
}

func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	params := mapper.MapBookingToSQLCCreate(booking)

	rawBooking, err := r.q.CreateBooking(ctx, db, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, pkgerrs.NewObjectAlreadyExistsErrorWithReason(
					"booking", pgErr,
				)
			}
		}
		return nil, err
	}

	return mapper.MapSQLCToBooking(rawBooking), nil
}

func (r *BookingRepository) Get(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawBooking, err := r.q.GetBookingByID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("booking", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToBooking(rawBooking), nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	return r.q.UpdateBookingStatus(ctx, db, sqlc.UpdateBookingStatusParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Status: status.String(),
	})
}

func (r *BookingRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawBookings, err := r.q.ListBookingsByUserID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToBookingsList(rawBookings), nil
}

func (r *BookingRepository) ListAll(ctx context.Context, limit, offset int32) ([]*model.Booking, int64, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	raw, err := r.q.ListAllBookings(
		ctx,
		db,
		sqlc.ListAllBookingsParams{
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		return nil, 0, err
	}

	mapped, total := mapper.MapSQLCAllToBookingsList(raw)

	return mapped, total, nil
}

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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	q      *sqlc.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewUserRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *UserRepository {
	return &UserRepository{
		q:      sqlc.New(),
		pool:   pgClient.Pool,
		getter: getter,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	params := mapper.MapUserToSQLCCreate(user)

	rawUser, err := r.q.CreateUser(ctx, db, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, pkgerrs.NewObjectAlreadyExistsErrorWithReason(
					"user", pgErr,
				)
			}
		}
		return nil, err
	}

	return mapper.MapSQLCToUser(rawUser), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawUser, err := r.q.GetUserByID(
		ctx,
		db,
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("user", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToUser(rawUser), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	rawUser, err := r.q.GetUserByEmail(ctx, db, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("user", email)
		}
		return nil, err
	}

	return mapper.MapSQLCToUser(rawUser), nil
}

func (r *UserRepository) EnsureDummyUsers(ctx context.Context, adminID, userID uuid.UUID) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)

	params := sqlc.EnsureDummyUsersParams{
		AdminID: pgtype.UUID{
			Bytes: adminID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	}
	return r.q.EnsureDummyUsers(ctx, db, params)
}

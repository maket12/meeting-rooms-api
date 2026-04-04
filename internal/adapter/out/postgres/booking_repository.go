package postgres

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	pkgpostgres "MeetingRoomsAPI/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
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

//func (r *BookingRepository) Create(ctx context.Context) {
//	return r.q.Creat
//}

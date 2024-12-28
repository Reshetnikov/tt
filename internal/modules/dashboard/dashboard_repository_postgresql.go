package dashboard

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type DashboardRepositoryPostgres struct {
	// db *pgxpool.Pool
	db PgxPool
}

func NewDashboardRepositoryPostgres(db PgxPool) *DashboardRepositoryPostgres {
	return &DashboardRepositoryPostgres{db: db}
}

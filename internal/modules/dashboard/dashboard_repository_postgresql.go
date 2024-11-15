package dashboard

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardRepositoryPostgres struct {
	db *pgxpool.Pool
}

func NewDashboardRepositoryPostgres(db *pgxpool.Pool) *DashboardRepositoryPostgres {
	return &DashboardRepositoryPostgres{db: db}
}

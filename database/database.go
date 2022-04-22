package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Connection string `yaml:"postgresql"`
}

type Database struct {
	db *sqlx.DB
}

func NewDatabase(ctx context.Context, cfg *Config) (*Database, error) {
	drv := stdlib.GetDefaultDriver().(*stdlib.Driver)

	ctor, err := drv.OpenConnector(cfg.Connection)
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(sql.OpenDB(ctor), "pgx")

	return &Database{dbx}, nil
}

func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *Database) Close(ctx context.Context) error {
	return d.db.Close()
}

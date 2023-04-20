package store

import (
	"context"
	"database/sql"
	"fmt"

	sqlc "github.com/Exca-DK/node-util/service/db/sqlc"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	Source        string
	MigrationPath string
}

type AtomicAction func(any) error

type Store interface {
	sqlc.Querier
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	*sqlc.Queries
	db *sql.DB
}

func NewStore(cfg Config) (Store, error) {
	if err := runMigration(cfg.MigrationPath, cfg.Source); err != nil {
		return nil, err
	}

	conn, err := sql.Open("postgres", cfg.Source)
	if err != nil {
		return nil, err
	}

	return &SQLStore{
		Queries: sqlc.New(conn),
		db:      conn,
	}, nil
}

// Executes atomic transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*sqlc.Queries) (any, error), after AtomicAction) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	obj, err := fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	if after != nil {
		if err := after(obj); err != nil {
			return fmt.Errorf("after action failure. error: %v", err)
		}
	}

	return tx.Commit()
}

func runMigration(migrationURL string, dbSource string) error {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		return fmt.Errorf("migrration init failure. err: %v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrration up failure. err: %v", err)
	}

	return nil
}

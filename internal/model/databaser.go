package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/humanitec/golib/hlogger"
	"github.com/lib/pq"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file" // File source driver for migrations

	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

// Model is the underlying type for the entire model.
type databaser struct {
	*sql.DB
	logger *hlogger.HLogger
}

// NewDatabaser creates a new database provider instance
func NewDatabaser(ctx context.Context, logger *hlogger.HLogger, connStr string, connectRetries int, runMigrations bool) (Databaser, error) {
	sqltrace.Register("postgres", &pq.Driver{})
	pg, err := sqltrace.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db := databaser{
		pg,
		logger,
	}

	// (Optional) Wait and connect
	if connectRetries > 0 {
		if err := db.Connect(ctx, connectRetries); err != nil {
			return &db, err
		}
	}

	// (Optional) Run migrations
	if runMigrations {
		if err := db.RunMigrations(); err != nil {
			return &db, err
		}
	}

	return &db, err
}

// sleepWithContext waits for a delay or until the ctx is cancelled.
// Returns an error if the ctx has been cancelled.
func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}

	return ctx.Err()
}

// Connect tries to run a query on the database ad if it fails, tries again after an increasing wait.
// This is useful because when deploying, it is often the case that the database only becomes accessible some time
// after deployment. (e.g. because the CloudSQL proxy takes time to establish the connection to the database.)
func (db *databaser) Connect(ctx context.Context, retries int) error {
	if retries <= 0 {
		return errors.New("connection retries counter should be greater than 0 (zero)")
	}

	var delay int
	var err error
	for attempt := 1; attempt <= retries; attempt++ {
		if err != nil {
			delay = int(1 << uint(attempt))
			db.logger.Logger.Sugar().Infof("Unable to connect to the database (%d/%d): %v. Waiting %d seconds...", attempt, retries, err, delay)
			if err = sleepWithContext(ctx, time.Duration(delay)*time.Second); err != nil {
				break
			}
		}

		db.logger.Logger.Sugar().Infof("Connecting to the database (%d/%d)...", attempt, retries)
		if _, err = db.DB.ExecContext(ctx, "SET timezone = 'utc'"); err == nil {
			return nil
		}
	}

	return fmt.Errorf("connecting to the database: %w", err)
}

// RunMigrations applies database migration scripts
func (db *databaser) RunMigrations() error {
	driver, _ := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("preparing migrations: %w", err)
	}

	db.logger.Logger.Sugar().Info("Applying database migration scripts...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if errors.Is(err, os.ErrNotExist) {
			// ignore if current migration doesn't exist, could be a result of the code rollback
			db.logger.Logger.Sugar().Warnw("Unable to apply database migration scripts", "err", err)
			return nil
		}
		return fmt.Errorf("applying migrations: %w", err)
	}

	return nil
}

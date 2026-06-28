package testhelpers

import (
	"backend/migrations"
	pkgpostgres "backend/pkg/postgres"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container *container.PostgresContainer
	Config    *pkgpostgres.Config
}

func StartPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	var (
		user     = "user"
		password = "password"
		dbName   = "testdb"
	)

	pgContainer, err := container.Run(ctx,
		"postgres:15-alpine",
		container.WithUsername(user),
		container.WithPassword(password),
		container.WithDatabase(dbName),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil || host == "" {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	cfg := pkgpostgres.NewConfig(
		host, int(port.Num()), user,
		password, dbName, "disable",
		10, 5,
		time.Minute, time.Minute,
	)

	return &PostgresContainer{
		Container: pgContainer,
		Config:    cfg,
	}, nil
}

func (pc *PostgresContainer) MigrateUp(version uint) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, pc.Config.MigrationDSN())
	if err != nil {
		return fmt.Errorf("failed to init migration tool: %w", err)
	}

	err = m.Migrate(version)

	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	var dirtyErr migrate.ErrDirty
	if errors.As(err, &dirtyErr) {
		_ = m.Force(dirtyErr.Version)
		_ = m.Down()
		err = m.Migrate(version)
		if err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return nil
}

func (pc *PostgresContainer) MigrateDown() error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, pc.Config.MigrationDSN())
	if err != nil {
		return fmt.Errorf("failed to init migration tool: %w", err)
	}

	return m.Down()
}

func (pc *PostgresContainer) Close(ctx context.Context) error {
	return pc.Container.Terminate(ctx)
}

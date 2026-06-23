package postgres_test

import (
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/domain/model"
	"backend/migrations"
	pkgerrs "backend/pkg/errs"
	pkgpostgres "backend/pkg/postgres"
	"backend/pkg/utils"
	"context"
	"errors"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RoomRepoSuite struct {
	suite.Suite
	dbClient *pkgpostgres.Client
	repo     *adapterpostgres.RoomRepository
	ctx      context.Context
	migrate  *migrate.Migrate
	testRoom *model.Room
}

func TestRoomRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RoomRepoSuite))
}

func (s *RoomRepoSuite) setupDatabase() {
	// Version of the lowest migration to apply
	const targetVersion = 2

	dbConfig := pkgpostgres.NewConfig(
		"localhost", 5433, "test-user",
		"test-pass", "test-db", "disable",
		5, 5,
		10*time.Second, 10*time.Second,
	)
	dsn := "postgres://test-user:test-pass@localhost:5433/test-db?sslmode=disable"

	dbClient, err := pkgpostgres.NewClient(context.Background(), dbConfig)
	s.Require().NoError(err)
	s.dbClient = dbClient

	sourceDriver, err := iofs.New(migrations.FS, ".")
	s.Require().NoError(err, "failed to create iofs driver")

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceDriver,
		dsn,
	)
	s.Require().NoError(err, "failed to create migration instance")

	s.migrate = m

	err = m.Migrate(targetVersion)

	// If migration is correct - setup has done
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return
	}

	// Except dirty db as a normal scenario
	var dirtyErr migrate.ErrDirty
	if !errors.As(err, &dirtyErr) {
		s.FailNowf("failed to migrate up", "unexpected error: %v", err)
	}

	// ================ Restore dirty database ================
	_ = m.Force(dirtyErr.Version)

	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "failed to migrate down during recovery")
	}

	err = m.Migrate(targetVersion)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "failed to migrate up after recovery")
	}
}

func (s *RoomRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.setupDatabase()
	s.repo = adapterpostgres.NewRoomRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	s.testRoom, _ = model.NewRoom(
		"Room №112",
		utils.VPtr("The very comfortable area with sofas and chairs"),
		utils.VPtr(100),
	)
}

func (s *RoomRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	s.dbClient.Close()
}

func (s *RoomRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx, "TRUNCATE TABLE rooms CASCADE")
	s.Require().NoError(err)
}

func (s *RoomRepoSuite) TestCreateGet() {
	// Create a test room at first
	_, err := s.repo.Create(s.ctx, s.testRoom)
	s.Require().NoError(err)

	// Then get it by id
	room, err := s.repo.Get(s.ctx, s.testRoom.ID())
	s.Require().NoError(err)
	s.Require().NotNil(room)
	s.Require().Exactly(s.testRoom.Description(), room.Description())
	s.Require().Exactly(s.testRoom.Capacity(), room.Capacity())
}

func (s *RoomRepoSuite) TestGet_NotFound() {
	// Try to get a non-existing room by id
	var unexistingID = uuid.New()
	room, err := s.repo.Get(s.ctx, unexistingID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(room)
}

func (s *RoomRepoSuite) TestList() {
	// Create rooms in advance
	_, _ = s.repo.Create(s.ctx, s.testRoom)

	// Get a list of them
	rooms, err := s.repo.List(s.ctx)

	s.Require().NoError(err)
	s.Require().NotNil(rooms)
	s.Require().Len(rooms, 1)
	s.Require().Exactly(s.testRoom.ID(), rooms[0].ID())
	s.Require().Exactly(s.testRoom.Name(), rooms[0].Name())
}

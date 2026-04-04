package postgres_test

import (
	adapterpostgres "MeetingRoomsAPI/internal/adapter/out/postgres"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/migrations"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	pkgpostgres "MeetingRoomsAPI/pkg/postgres"
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

type ScheduleRepoSuite struct {
	suite.Suite
	dbClient     *pkgpostgres.Client
	repo         *adapterpostgres.ScheduleRepository
	ctx          context.Context
	migrate      *migrate.Migrate
	testSchedule *model.Schedule
}

func TestScheduleRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(ScheduleRepoSuite))
}

func (s *ScheduleRepoSuite) setupDatabase() {
	// Version of the lowest migration to apply
	const targetVersion = 3

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

func (s *ScheduleRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.setupDatabase()
	s.repo = adapterpostgres.NewScheduleRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)

	// Create a room in the rooms table
	testRoom, _ := model.NewRoom("№100", nil, nil)

	roomsRepo := adapterpostgres.NewRoomRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	_, _ = roomsRepo.Create(s.ctx, testRoom)

	s.testSchedule, _ = model.NewSchedule(
		testRoom.ID(),
		[]int{1, 3, 5},
		"10:00",
		"11:00",
	)
}

func (s *ScheduleRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	s.dbClient.Close()
}

func (s *ScheduleRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx, "TRUNCATE TABLE schedules CASCADE")
	s.Require().NoError(err)
}

func (s *ScheduleRepoSuite) TestCreateGet() {
	// Create a test schedule at first
	err := s.repo.Create(s.ctx, s.testSchedule)
	s.Require().NoError(err)

	// Then get it by room id
	schedule, err := s.repo.Get(s.ctx, s.testSchedule.RoomID())
	s.Require().NoError(err)
	s.Require().NotNil(schedule)
	s.Require().Exactly(s.testSchedule.ID(), schedule.ID())
	s.Require().ElementsMatch(s.testSchedule.DaysOfWeek(), schedule.DaysOfWeek())
	s.Require().Exactly(s.testSchedule.StartTime(), schedule.StartTime())
	s.Require().Exactly(s.testSchedule.EndTime(), schedule.EndTime())
}

func (s *ScheduleRepoSuite) TestGet_NotFound() {
	// Try to get a non-existing schedule by room id
	var unexistingRoomID = uuid.New()
	schedule, err := s.repo.Get(s.ctx, unexistingRoomID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(schedule)
}

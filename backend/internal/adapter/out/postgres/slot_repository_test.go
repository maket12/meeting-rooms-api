package postgres_test

import (
	"backend/internal/adapter/out/postgres"
	model2 "backend/internal/domain/model"
	"backend/migrations"
	pkgerrs "backend/pkg/errs"
	pkgpostgres "backend/pkg/postgres"
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

type SlotRepoSuite struct {
	suite.Suite
	dbClient  *pkgpostgres.Client
	repo      *postgres.SlotRepository
	ctx       context.Context
	migrate   *migrate.Migrate
	testSlots []*model2.Slot
}

func TestSlotRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(SlotRepoSuite))
}

func (s *SlotRepoSuite) setupDatabase() {
	// Version of the lowest migration to apply
	const targetVersion = 5

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

func (s *SlotRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.setupDatabase()
	s.repo = postgres.NewSlotRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)

	// Create rooms in advance
	room1, _ := model2.NewRoom("212", nil, nil)
	room2, _ := model2.NewRoom("213", nil, nil)

	roomsRepo := postgres.NewRoomRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	_, _ = roomsRepo.Create(s.ctx, room1)
	_, _ = roomsRepo.Create(s.ctx, room2)

	// Now we can create slots
	startTime := time.Now().UTC()

	slot1, _ := model2.NewSlot(room1.ID(), startTime)
	slot2, _ := model2.NewSlot(room1.ID(), startTime.Add(30*time.Minute))
	slot3, _ := model2.NewSlot(room2.ID(), startTime)

	s.testSlots = []*model2.Slot{slot1, slot2, slot3}
}

func (s *SlotRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	s.dbClient.Close()
}

func (s *SlotRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx, "TRUNCATE TABLE slots CASCADE")
	s.Require().NoError(err)
}

func (s *SlotRepoSuite) TestCreateBatchGet() {
	// Create slots
	err := s.repo.CreateBatch(s.ctx, s.testSlots)
	s.Require().NoError(err)

	// Get them by their ids
	for _, testSlot := range s.testSlots {
		slot, err := s.repo.Get(s.ctx, testSlot.ID())
		s.Require().NoError(err)
		s.Require().NotNil(slot)

		s.Equal(testSlot.RoomID(), slot.RoomID())
		s.Equal(testSlot.Start().Round(time.Second), slot.Start().Round(time.Second))
		s.Equal(testSlot.End().Round(time.Second), slot.End().Round(time.Second))
	}
}

func (s *SlotRepoSuite) TestGet_NotFound() {
	// Try to get a non-existing slot by id
	var unexistingID = uuid.New()
	slot, err := s.repo.Get(s.ctx, unexistingID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(slot)
}

func (s *SlotRepoSuite) TestListFree() {
	// Create slots in advance
	_ = s.repo.CreateBatch(s.ctx, s.testSlots)

	// Case №1: get by room id of the first slot (expect 2 slots in result)
	slots, err := s.repo.ListFree(
		s.ctx,
		s.testSlots[0].RoomID(),
		s.testSlots[0].Start(),
	)
	s.Require().NoError(err)
	s.Require().NotNil(slots)
	s.Require().Len(slots, 2)

	// Case №2: get by room id of the third slot (expect 1 slot in result)
	slots, err = s.repo.ListFree(
		s.ctx,
		s.testSlots[2].RoomID(),
		s.testSlots[2].Start(),
	)
	s.Require().NoError(err)
	s.Require().NotNil(slots)
	s.Require().Len(slots, 1)

	// Case №2: specify too late date (expect 0 slots in result)
	slots, err = s.repo.ListFree(
		s.ctx,
		s.testSlots[0].RoomID(),
		s.testSlots[0].Start().Add(48*time.Hour),
	)
	s.Require().NoError(err)
	s.Require().NotNil(slots)
	s.Require().Empty(slots)
}

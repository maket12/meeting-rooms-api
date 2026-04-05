package postgres_test

import (
	adapterpostgres "MeetingRoomsAPI/internal/adapter/out/postgres"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/migrations"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	pkgpostgres "MeetingRoomsAPI/pkg/postgres"
	"MeetingRoomsAPI/pkg/utils"
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

type BookingRepoSuite struct {
	suite.Suite
	dbClient    *pkgpostgres.Client
	repo        *adapterpostgres.BookingRepository
	ctx         context.Context
	migrate     *migrate.Migrate
	testBooking *model.Booking
	testUserID  uuid.UUID
	testRoomID  uuid.UUID
	testSlotID  uuid.UUID
}

func TestBookingRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(BookingRepoSuite))
}

func (s *BookingRepoSuite) setupDatabase() {
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

func (s *BookingRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.setupDatabase()
	s.repo = adapterpostgres.NewBookingRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)

	userRepo := adapterpostgres.NewUserRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	testUser, _ := model.NewUser(
		"email",
		"hash",
		model.RoleUser,
	)
	_, _ = userRepo.Create(s.ctx, testUser)
	s.testUserID = testUser.ID()

	roomRepo := adapterpostgres.NewRoomRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	testRoom, _ := model.NewRoom("№147", nil, nil)
	_, _ = roomRepo.Create(s.ctx, testRoom)
	s.testRoomID = testRoom.ID()

	slotRepo := adapterpostgres.NewSlotRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	testSlot, _ := model.NewSlot(s.testRoomID, time.Now().Add(time.Hour).UTC())
	_ = slotRepo.CreateBatch(s.ctx, []*model.Slot{testSlot})
	s.testSlotID = testSlot.ID()

	s.testBooking, _ = model.NewBooking(
		s.testSlotID,
		s.testUserID,
		utils.VPtr("https://telemost.yandex.ru/test"),
	)
}

func (s *BookingRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	s.dbClient.Close()
}

func (s *BookingRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx, "TRUNCATE TABLE bookings CASCADE")
	s.Require().NoError(err)
}

func (s *BookingRepoSuite) TestCreateGet() {
	// Create a test booking at first
	_, err := s.repo.Create(s.ctx, s.testBooking)
	s.Require().NoError(err)

	// Then get it by id
	booking, err := s.repo.Get(s.ctx, s.testBooking.ID())
	s.Require().NoError(err)
	s.Require().NotNil(booking)
	s.Require().Equal(*s.testBooking.ConferenceLink(), *booking.ConferenceLink())
	s.Require().Exactly(s.testBooking.Status(), booking.Status())
}

func (s *BookingRepoSuite) TestGet_NotFound() {
	// Try to get a non-existing booking by id
	var unexistingID = uuid.New()
	booking, err := s.repo.Get(s.ctx, unexistingID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(booking)
}

func (s *BookingRepoSuite) TestUpdateStatus() {
	// Create the booking in advance
	_, _ = s.repo.Create(s.ctx, s.testBooking)

	// Cancel it
	err := s.repo.UpdateStatus(s.ctx, s.testBooking.ID(), model.BookingCancelled)
	s.Require().NoError(err)

	// Check the result
	booking, _ := s.repo.Get(s.ctx, s.testBooking.ID())
	s.Require().Equal(model.BookingCancelled, booking.Status())
}

func (s *BookingRepoSuite) TestListByUserID() {
	// Create in advance
	_, _ = s.repo.Create(s.ctx, s.testBooking)

	// Expect 1 item in result
	bookings, err := s.repo.ListByUserID(s.ctx, s.testUserID)
	s.Require().NoError(err)
	s.Require().Len(bookings, 1)
	s.Require().Equal(s.testBooking.ID(), bookings[0].ID())
}

func (s *BookingRepoSuite) TestListAll() {
	// Create in advance
	_, _ = s.repo.Create(s.ctx, s.testBooking)

	// Expect 1 item in result
	bookings, total, err := s.repo.ListAll(s.ctx, 10, 0)
	s.Require().NoError(err)
	s.Require().Equal(int64(1), total)
	s.Require().Len(bookings, 1)
}

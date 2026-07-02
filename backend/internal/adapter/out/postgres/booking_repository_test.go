//go:build integration

package postgres_test

import (
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/domain/model"
	pkgerrs "backend/pkg/errs"
	"backend/pkg/utils"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type BookingRepoSuite struct {
	BaseRepoSuite
	repo        *adapterpostgres.BookingRepository
	testBooking *model.Booking
	testUserID  uuid.UUID
	testRoomID  uuid.UUID
	testSlotID  uuid.UUID
}

func TestBookingRepoSuite(t *testing.T) { suite.Run(t, new(BookingRepoSuite)) }

func (s *BookingRepoSuite) SetupSuite() {
	s.SetupBase(5)
	s.repo = adapterpostgres.NewBookingRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
}

func (s *BookingRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "bookings", "slots", "rooms", "users")
	s.Require().NoError(err)

	s.seedData()
}

func (s *BookingRepoSuite) seedData() {
	userRepo := adapterpostgres.NewUserRepository(s.dbClient, trmpgx.DefaultCtxGetter)
	roomRepo := adapterpostgres.NewRoomRepository(s.dbClient, trmpgx.DefaultCtxGetter)
	slotRepo := adapterpostgres.NewSlotRepository(s.dbClient, trmpgx.DefaultCtxGetter)

	testUser, _ := model.NewUser("email", "hash", model.RoleUser.String())
	testRoom, _ := model.NewRoom("№147", nil, nil)
	testSlot, _ := model.NewSlot(testRoom.ID(), time.Now().Add(time.Hour).UTC())

	_, err := userRepo.Create(s.ctx, testUser)
	s.Require().NoError(err, "failed to seed user")

	_, err = roomRepo.Create(s.ctx, testRoom)
	s.Require().NoError(err, "failed to seed room")

	_ = slotRepo.CreateBatch(s.ctx, []*model.Slot{testSlot})
	s.Require().NoError(err, "failed to seed slots")

	s.testUserID = testUser.ID()
	s.testRoomID = testRoom.ID()
	s.testSlotID = testSlot.ID()

	s.testBooking, _ = model.NewBooking(s.testSlotID, s.testUserID, utils.VPtr("https://telemost.yandex.ru/test"))
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

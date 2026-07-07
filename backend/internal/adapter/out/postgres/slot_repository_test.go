//go:build integration

package postgres_test

import (
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/domain/model"
	pkgerrs "backend/pkg/errs"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

var startTime = time.Now().Add(time.Hour).UTC()

type SlotRepoSuite struct {
	BaseRepoSuite
	repo      *adapterpostgres.SlotRepository
	testSlots []*model.Slot
}

func TestSlotRepoSuite(t *testing.T) { suite.Run(t, new(SlotRepoSuite)) }

func (s *SlotRepoSuite) SetupSuite() {
	s.SetupBase(5)
	s.repo = adapterpostgres.NewSlotRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
}

func (s *SlotRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "slots", "rooms")
	s.Require().NoError(err)

	s.seedData()
}

func (s *SlotRepoSuite) seedData() {
	roomsRepo := adapterpostgres.NewRoomRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)

	// Create rooms and slots in advance
	room1, _ := model.NewRoom("212", nil, nil)
	room2, _ := model.NewRoom("213", nil, nil)
	slot1, _ := model.NewSlot(room1.ID(), startTime)
	slot2, _ := model.NewSlot(room1.ID(), startTime.Add(30*time.Minute))
	slot3, _ := model.NewSlot(room2.ID(), startTime)

	_, err := roomsRepo.Create(s.ctx, room1)
	s.Require().NoError(err, "failed to seed room")

	_, err = roomsRepo.Create(s.ctx, room2)
	s.Require().NoError(err, "failed to seed room")

	s.testSlots = []*model.Slot{slot1, slot2, slot3}
}

func (s *SlotRepoSuite) TestCreateBatchGet() {
	// Create slots
	err := s.repo.CreateBatch(s.ctx, s.testSlots)
	s.Require().NoError(err)

	// Get them by their ids
	for _, testSlot := range s.testSlots {
		slot, getErr := s.repo.Get(s.ctx, testSlot.ID())
		s.Require().NoError(getErr)
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

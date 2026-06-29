//go:build integration

package postgres_test

import (
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/domain/model"
	pkgerrs "backend/pkg/errs"
	"testing"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ScheduleRepoSuite struct {
	BaseRepoSuite
	repo         *adapterpostgres.ScheduleRepository
	testSchedule *model.Schedule
}

func TestScheduleRepoSuite(t *testing.T) { suite.Run(t, new(ScheduleRepoSuite)) }

func (s *ScheduleRepoSuite) SetupSuite() {
	s.SetupBase(3)
	s.repo = adapterpostgres.NewScheduleRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
}

func (s *ScheduleRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx,
		"TRUNCATE TABLE schedules, rooms RESTART IDENTITY CASCADE",
	)
	s.Require().NoError(err)

	s.seedData()
}

func (s *ScheduleRepoSuite) seedData() {
	roomsRepo := adapterpostgres.NewRoomRepository(s.dbClient, trmpgx.DefaultCtxGetter)

	// Create a room in the rooms table
	testRoom, _ := model.NewRoom("№100", nil, nil)

	_, err := roomsRepo.Create(s.ctx, testRoom)
	s.Require().NoError(err, "failed to seed room")

	s.testSchedule, _ = model.NewSchedule(testRoom.ID(), []int{1, 3, 5}, "10:00", "11:00")
}

func (s *ScheduleRepoSuite) TestCreateGet() {
	// Create a test schedule at first
	schedule, err := s.repo.Create(s.ctx, s.testSchedule)
	s.Require().NoError(err)
	s.Require().NotNil(schedule)

	// Then get it by room id
	schedule, err = s.repo.Get(s.ctx, s.testSchedule.RoomID())
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

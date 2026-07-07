//go:build integration

package postgres_test

import (
	"testing"

	adapterpostgres "github.com/maket12/meeting-rooms-api/internal/adapter/out/postgres"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
	"github.com/maket12/meeting-rooms-api/pkg/utils"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RoomRepoSuite struct {
	BaseRepoSuite
	repo     *adapterpostgres.RoomRepository
	testRoom *model.Room
}

func TestRoomRepoSuite(t *testing.T) {
	suite.Run(t, new(RoomRepoSuite))
}

func (s *RoomRepoSuite) SetupSuite() {
	s.SetupBase(2)
	s.repo = adapterpostgres.NewRoomRepository(s.dbClient, trmpgx.DefaultCtxGetter)
	s.testRoom, _ = model.NewRoom(
		"Room №112",
		utils.VPtr("The very comfortable area with sofas and chairs"),
		utils.VPtr(100),
	)
}

func (s *RoomRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "rooms")
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

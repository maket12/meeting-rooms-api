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

type UserRepoSuite struct {
	BaseRepoSuite
	repo     *adapterpostgres.UserRepository
	testUser *model.User
}

func TestUserRepoSuite(t *testing.T) { suite.Run(t, new(UserRepoSuite)) }

func (s *UserRepoSuite) SetupSuite() {
	s.SetupBase(1)
	s.repo = adapterpostgres.NewUserRepository(
		s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	s.testUser, _ = model.NewUser(
		"test-email@avito.ru",
		"hashed_password",
		"user",
	)
}

func (s *UserRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "users")
	s.Require().NoError(err)
}

func (s *UserRepoSuite) TestCreateGet() {
	// Create a test user at first
	_, err := s.repo.Create(s.ctx, s.testUser)
	s.Require().NoError(err)

	// Then get him by id
	user, err := s.repo.GetByID(s.ctx, s.testUser.ID())
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.Require().Exactly(s.testUser.Email(), user.Email())
	s.Require().Exactly(s.testUser.Role(), user.Role())
	s.Require().WithinDuration(s.testUser.CreatedAt(), user.CreatedAt(), time.Second)

	// And get him by email
	user, err = s.repo.GetByEmail(s.ctx, s.testUser.Email())
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.Require().Exactly(s.testUser.ID(), user.ID())
	s.Require().Exactly(s.testUser.PasswordHash(), user.PasswordHash())
	s.Require().Exactly(s.testUser.Role(), user.Role())
}

func (s *UserRepoSuite) TestCreate_UniqueViolation() {
	// Create a test user at first
	_, _ = s.repo.Create(s.ctx, s.testUser)

	// Try to create a user with the same id
	user, err := s.repo.Create(s.ctx, s.testUser)
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectAlreadyExists)
	s.Require().Nil(user)

	// Try to create a user with the same email
	newUser, _ := model.NewUser(
		s.testUser.Email(),
		s.testUser.PasswordHash(),
		s.testUser.Role().String(),
	)
	user, err = s.repo.Create(s.ctx, newUser)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectAlreadyExists)
	s.Require().Nil(user)
}

func (s *UserRepoSuite) TestGet_NotFound() {
	// Try to get a non-existing user by id
	var unexistingID = uuid.New()
	user, err := s.repo.GetByID(s.ctx, unexistingID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(user)

	// Try to get a non-existing user by email
	var unexistingEmail = "not-exist@avito.ru"
	user, err = s.repo.GetByEmail(s.ctx, unexistingEmail)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(user)
}

func (s *UserRepoSuite) TestEnsureDummyUsers() {
	var (
		testAdminID = uuid.New()
		testUserID  = uuid.New()
	)

	// Call the method
	err := s.repo.EnsureDummyUsers(s.ctx, testAdminID, testUserID)
	s.Require().NoError(err)

	// Expect users to be created
	user, _ := s.repo.GetByID(s.ctx, testAdminID)
	s.Require().NotNil(user)

	user, _ = s.repo.GetByID(s.ctx, testUserID)
	s.Require().NotNil(user)
}

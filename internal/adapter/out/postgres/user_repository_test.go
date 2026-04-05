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

type UserRepoSuite struct {
	suite.Suite
	dbClient *pkgpostgres.Client
	repo     *adapterpostgres.UserRepository
	ctx      context.Context
	migrate  *migrate.Migrate
	testUser *model.User
}

func TestUserRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(UserRepoSuite))
}

func (s *UserRepoSuite) setupDatabase() {
	// Version of the lowest migration to apply
	const targetVersion = 1

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

func (s *UserRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.setupDatabase()
	s.repo = adapterpostgres.NewUserRepository(s.dbClient, trmpgx.DefaultCtxGetter)
	s.testUser, _ = model.NewUser(
		"test-email@avito.ru",
		"hashed_password",
		"user",
	)
}

func (s *UserRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	s.dbClient.Close()
}

func (s *UserRepoSuite) SetupTest() {
	_, err := s.dbClient.Pool.Exec(s.ctx, "TRUNCATE TABLE users CASCADE")
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
		s.testUser.Role(),
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

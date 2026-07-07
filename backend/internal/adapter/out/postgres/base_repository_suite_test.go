//go:build integration

package postgres_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/maket12/meeting-rooms-api/internal/testhelpers"
	pkgpostgres "github.com/maket12/meeting-rooms-api/pkg/postgres"

	"github.com/stretchr/testify/suite"
)

var (
	globalContainer *testhelpers.PostgresContainer
	globalClient    *pkgpostgres.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Init postgres container
	pgContainer, err := testhelpers.StartPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Could not start postgres: %v", err)
	}
	globalContainer = pgContainer

	// Init postgres client
	client, err := pkgpostgres.NewClient(ctx, pgContainer.Config)
	if err != nil {
		log.Fatalf("Could not connect to postgres: %v", err)
	}
	globalClient = client

	// Launch all tests
	code := m.Run()

	// Delete container
	globalClient.Close()
	_ = globalContainer.Close(ctx)

	os.Exit(code)
}

type BaseRepoSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	dbClient    *pkgpostgres.Client
	ctx         context.Context
}

func (s *BaseRepoSuite) SetupBase(version uint) {
	s.pgContainer = globalContainer
	s.dbClient = globalClient
	s.ctx = context.Background()

	// Apply migrations
	err := s.pgContainer.MigrateUp(version)
	s.Require().NoError(err)
}

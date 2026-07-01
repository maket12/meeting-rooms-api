///go:build e2e

package e2e

import (
	adapterconf "backend/internal/adapter/out/conference"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"backend/cmd/app/config"
	adapterhttp "backend/internal/adapter/in/http"
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/app/usecase"
	infrajwt "backend/internal/infrastructure/jwt"
	infrapasswd "backend/internal/infrastructure/password"
	"backend/internal/testhelpers"
	pkgpostgres "backend/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	migrationVersion = 5
	apiVersion       = "v1"
)

type testApp struct {
	server     *httptest.Server
	client     *http.Client
	pg         *testhelpers.PostgresContainer
	dbClient   *pkgpostgres.Client
	cfg        *config.TestConfig
	userToken  *string
	adminToken *string
}

var (
	appInstance *testApp
	once        sync.Once
)

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newLogger(level string) *slog.Logger {
	logLevel := parseLogLevel(level)
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func newPostgresClient(ctx context.Context, cfg *config.TestConfig) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
		cfg.DBName, cfg.DbSSLMode, cfg.DbMaxConn,
		cfg.DbMinConn, cfg.DbMaxConnLifeTime, cfg.DbMaxConnIdleTime,
	)

	return pkgpostgres.NewClient(ctx, pgConfig)
}

func setupE2E(t *testing.T) *testApp {
	once.Do(func() {
		ctx := context.Background()

		cfg, err := config.LoadTest()
		require.NoError(t, err)

		pg, err := testhelpers.StartPostgresContainer(ctx)
		require.NoError(t, err)

		err = pg.MigrateUp(migrationVersion)
		require.NoError(t, err)

		cfg.DbUser = pg.Config.User
		cfg.DbPassword = pg.Config.Password
		cfg.DBName = pg.Config.Name
		cfg.DbHost = pg.Config.Host
		cfg.DbPort = pg.Config.Port

		logger := newLogger(cfg.LogLevel)

		pgClient, err := newPostgresClient(ctx, cfg)
		require.NoError(t, err)

		trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

		// Repositories
		userRepo := adapterpostgres.NewUserRepository(pgClient, trmpgx.DefaultCtxGetter)
		roomRepo := adapterpostgres.NewRoomRepository(pgClient, trmpgx.DefaultCtxGetter)
		scheduleRepo := adapterpostgres.NewScheduleRepository(pgClient, trmpgx.DefaultCtxGetter)
		slotRepo := adapterpostgres.NewSlotRepository(pgClient, trmpgx.DefaultCtxGetter)
		bookingRepo := adapterpostgres.NewBookingRepository(pgClient, trmpgx.DefaultCtxGetter)
		conferenceService := adapterconf.NewConferenceService("available")
		jwtGen := infrajwt.NewTokenGenerator(cfg.AuthSecret, cfg.AuthTTL)
		passHasher := infrapasswd.NewHasher(cfg.PasswordCost)

		// Adding seed-data
		err = userRepo.EnsureDummyUsers(ctx, cfg.DummyAdminID, cfg.DummyUserID)
		require.NoError(t, err)

		dummyLoginUC := usecase.NewDummyLoginUC(
			userRepo, jwtGen,
			cfg.DummyAdminID, cfg.DummyUserID,
		)
		registerUC := usecase.NewRegisterUC(userRepo, passHasher)
		loginUC := usecase.NewLoginUC(userRepo, passHasher, jwtGen)
		createRoomUC := usecase.NewCreateRoomUC(roomRepo)
		listRoomsUC := usecase.NewListRoomsUC(roomRepo)
		createScheduleUC := usecase.NewCreateScheduleUC(
			trManager, roomRepo, scheduleRepo, slotRepo,
		)
		listSlotsUC := usecase.NewListSlotsUC(
			trManager, roomRepo, scheduleRepo, slotRepo,
		)
		createBookingUC := usecase.NewCreateBookingUC(
			trManager, slotRepo,
			bookingRepo, conferenceService,
		)
		cancelBookingUC := usecase.NewCancelBookingUC(bookingRepo)
		listMyBookingsUC := usecase.NewListMyBookingsUC(bookingRepo)
		listBookingsUC := usecase.NewListBookingsUC(bookingRepo)

		authHandler := adapterhttp.NewAuthHandler(
			logger,
			dummyLoginUC,
			registerUC,
			loginUC,
		)
		roomHandler := adapterhttp.NewRoomHandler(
			logger,
			createRoomUC,
			listRoomsUC,
		)
		scheduleHandler := adapterhttp.NewScheduleHandler(logger, createScheduleUC)
		slotHandler := adapterhttp.NewSlotHandler(logger, listSlotsUC)
		bookingHandler := adapterhttp.NewBookingHandler(
			logger,
			createBookingUC,
			cancelBookingUC,
			listMyBookingsUC,
			listBookingsUC,
		)

		router := adapterhttp.NewRouter(
			authHandler,
			roomHandler,
			scheduleHandler,
			slotHandler,
			bookingHandler,
			jwtGen,
		).InitRoutes(logger)

		ts := httptest.NewServer(router)

		appInstance = &testApp{
			server:   ts,
			client:   ts.Client(),
			pg:       pg,
			dbClient: pgClient,
			cfg:      cfg,
		}
	})

	appInstance.cleanData(t, context.Background())

	return appInstance
}

// cleanData Uses migration tools to reset the database in its pure form
// and provides the necessary seed data.
func (a *testApp) cleanData(t *testing.T, ctx context.Context) {
	err := a.pg.TruncateTables(ctx)
	require.NoError(t, err, "failed to truncate tables")

	userRepo := adapterpostgres.NewUserRepository(a.dbClient, trmpgx.DefaultCtxGetter)

	// Adding seed-data
	err = userRepo.EnsureDummyUsers(ctx, a.cfg.DummyAdminID, a.cfg.DummyUserID)
	require.NoError(t, err, "failed to re-seed data")
}

func (a *testApp) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, a.server.URL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.client.Do(req)
}

func (a *testApp) makeRequestAuth(method, path string, body interface{}, token string) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req, _ := http.NewRequest(method, a.server.URL+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.client.Do(req)
}

func (a *testApp) getToken(t *testing.T, role string) string {
	availableRoles := [2]string{"user", "admin"}
	require.Contains(t, availableRoles, role, "this role is not supported")

	resp, err := a.makeRequest(http.MethodPost, "/dummyLogin", map[string]interface{}{"role": role})
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)

	require.NoError(t, err)
	require.NotEmpty(t, body["token"])

	token, ok := body["token"].(string)
	require.True(t, ok)

	return token
}

// Helper for e2e tests.
// Returns the user jwt token if it's available.
// Otherwise, make request to the "/dummyLogin" endpoint to claim it.
func (a *testApp) getUserToken(t *testing.T) string {
	if a.userToken != nil {
		return *a.userToken
	}

	token := a.getToken(t, "user")
	a.userToken = &token

	return token
}

// Helper for e2e tests.
// Returns the admin jwt token if it's available.
// Otherwise, make request to the "/dummyLogin" endpoint to claim it.
func (a *testApp) getAdminToken(t *testing.T) string {
	if a.adminToken != nil {
		return *a.adminToken
	}

	token := a.getToken(t, "admin")
	a.adminToken = &token

	return token
}

// Helper for e2e tests.
// Creates the new user with specified parameters.  **
// Returns the id of the created user.
// ** If parameters are not specified, then it uses default values instead.
func (a *testApp) createUser(t *testing.T, email, password, role *string) string {
	const path = "/register"

	var userEmail, userPass, userRole string

	if email != nil {
		userEmail = *email
	} else {
		userEmail = "test123@gmail.com"
	}

	if password != nil {
		userPass = *password
	} else {
		userPass = "test-pass-123"
	}

	if role != nil {
		userRole = *role
	} else {
		userRole = "admin"
	}

	payload := map[string]interface{}{
		"email":    userEmail,
		"password": userPass,
		"role":     userRole,
	}

	resp, err := a.makeRequest(http.MethodPost, path, payload)
	require.NoError(t, err)

	var user map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	id := user["user"]["id"].(string)
	assert.NotEmpty(t, id)

	_ = resp.Body.Close()

	return id
}

func (a *testApp) deleteLocation(t *testing.T, slug string) {
	path := fmt.Sprintf("/api/v1/admin/locations/%s", slug)
	resp, err := a.makeRequestAuth("DELETE", path, nil, a.getAdminToken(t))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func (a *testApp) deactivateLocation(t *testing.T, slug string) {
	resp, err := a.makeRequestAuth(
		"PATCH",
		fmt.Sprintf("/api/v1/admin/locations/%s", slug),
		map[string]interface{}{"is_active": false},
		a.getAdminToken(t),
	)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

// Helper for e2e tests.
// Creates the new item with default values.
// Returns external id of the created item.
func (a *testApp) createItem(t *testing.T, payload map[string]interface{}) string {
	const path = "/api/v1/admin/items"

	if payload == nil {
		payload = map[string]interface{}{
			"name":        "Сэндвич с курицей",
			"description": "Сэндвич с курицей и соусом тар-тар",
			"category":    "breakfast",
			"photo_url":   "https://photos-storage/exsa129csa7690/chicken_sandwich.png",
			"nutrition": map[string]interface{}{
				"calories": 200,
				"proteins": 23.6,
				"fats":     1.9,
				"carbs":    0.3,
			},
		}
	}

	resp, err := a.makeRequestAuth(http.MethodPost, path, payload, a.getAdminToken(t))
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	var item map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&item)
	require.NoError(t, err)

	idStr := item["item"]["id"].(string)

	_, err = uuid.Parse(idStr)
	require.NoError(t, err)

	return idStr
}

func (a *testApp) deleteItem(t *testing.T, itemID string) {
	resp, err := a.makeRequestAuth(
		"DELETE",
		fmt.Sprintf("/api/v1/admin/items/%s", itemID),
		nil,
		a.getAdminToken(t),
	)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

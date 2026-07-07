//go:build e2e

package e2e

import (
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
	"time"

	adapterconf "github.com/maket12/meeting-rooms-api/internal/adapter/out/conference"

	"github.com/maket12/meeting-rooms-api/cmd/app/config"
	adapterhttp "github.com/maket12/meeting-rooms-api/internal/adapter/in/http"
	adapterpostgres "github.com/maket12/meeting-rooms-api/internal/adapter/out/postgres"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	infrajwt "github.com/maket12/meeting-rooms-api/internal/infrastructure/jwt"
	infrapasswd "github.com/maket12/meeting-rooms-api/internal/infrastructure/password"
	"github.com/maket12/meeting-rooms-api/internal/testhelpers"
	pkgpostgres "github.com/maket12/meeting-rooms-api/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationVersion = 5

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
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer func() { _ = resp.Body.Close() }()

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
// Returns the id of the created user and its jwt token.
// ** If parameters are not specified, then it uses default values instead.
func (a *testApp) createUser(t *testing.T, email, password, role *string) (string, string) {
	const (
		registerPath = "/register"
		loginPath    = "/login"
	)

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

	// Make request on "/register" endpoint to create a user firstly
	payload := map[string]interface{}{
		"email":    userEmail,
		"password": userPass,
		"role":     userRole,
	}

	resp, err := a.makeRequest(http.MethodPost, registerPath, payload)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	var user map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	id := user["user"]["id"].(string)
	assert.NotEmpty(t, id)

	// Now make request on "/login" endpoint to get access token
	payload = map[string]interface{}{"email": userEmail, "password": userPass}

	resp, err = a.makeRequest(http.MethodPost, loginPath, payload)
	require.NoError(t, err)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)

	require.NoError(t, err)
	require.NotEmpty(t, body["token"])

	token, ok := body["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	return id, token
}

// Helper for e2e tests.
// Creates the new room with default values.
// Returns external id of the created room.
func (a *testApp) createRoom(t *testing.T) string {
	const path = "/rooms/create"

	payload := map[string]interface{}{"name": "B122", "capacity": 50}

	resp, err := a.makeRequestAuth(http.MethodPost, path, payload, a.getAdminToken(t))
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	var room map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&room)
	require.NoError(t, err)

	id := room["room"]["id"].(string)

	_, err = uuid.Parse(id)
	require.NoError(t, err)

	return id
}

// Helper for e2e tests.
// Creates the new schedule with default values for specified room.
// Make sure to create the room in advance before calling this method.
func (a *testApp) createSchedule(t *testing.T, roomID string) {
	path := fmt.Sprintf("/rooms/%s/schedule/create", roomID)
	payload := map[string]interface{}{
		"days_of_week": []int{1, 2, 3, 4, 5, 6, 7},
		"start_time":   "8:30",
		"end_time":     "9:30",
	}

	resp, err := a.makeRequestAuth(http.MethodPost, path, payload, a.getAdminToken(t))
	require.NoError(t, err)

	_ = resp.Body.Close()
}

// Helper for e2e tests.
// Requests slots for specified room within next day trough "../slots/list" endpoint.
// Make sure you created schedule for the room before to call this method.
func (a *testApp) getSlots(t *testing.T, roomID string) []map[string]interface{} {
	testDate := time.Now().Add(24 * time.Hour).Format(time.DateOnly)

	path := fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, testDate)

	resp, err := a.makeRequestAuth(http.MethodGet, path, nil, a.getUserToken(t))
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	var slots map[string][]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&slots)
	require.NoError(t, err)

	assert.NotEmpty(t, slots["slots"])

	return slots["slots"]
}

// Helper for e2e tests.
// Creates the new booking for specified slot.
// Make sure to generate slots (e.g. create a schedule)
// in advance before calling this method.
// Returns id of created booking
func (a *testApp) createBooking(t *testing.T, slotID string) string {
	const path = "/bookings/create"

	payload := map[string]interface{}{"slot_id": slotID, "create_conference_link": false}

	resp, err := a.makeRequestAuth(http.MethodPost, path, payload, a.getUserToken(t))
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	var body map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	id := body["booking"]["id"].(string)
	require.NotEmpty(t, id)

	_, err = uuid.Parse(id)
	require.NoError(t, err)

	return id
}

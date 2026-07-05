///go:build load

package load

import (
	adapterconf "backend/internal/adapter/out/conference"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

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
	"golang.org/x/time/rate"
)

const (
	migrationVersion = 5
	apiVersion       = "v1"
)

const (
	seedRoomsCount         = 50
	seedSlotsPerRoomPerDay = 20 // 50 * 20 = 1000 slots per day
	seedUsersCount         = 500
	seedHistoricBookings   = 100_000
	freshSlotsForBooking   = 20_000

	targetRPS          = 100
	targetSuccessRate  = 99.9 // %
	slotsListP99Target = 200 * time.Millisecond
)

type testApp struct {
	server     *httptest.Server
	client     *http.Client
	pg         *testhelpers.PostgresContainer
	dbClient   *pkgpostgres.Client
	cfg        *config.TestConfig
	userToken  *string
	adminToken *string

	roomIDs   []string
	userPool  []poolUser
	freeSlots chan string
}

type poolUser struct {
	id    string
	token string
}

var (
	appInstance *testApp
	once        sync.Once
)

func newPostgresClient(ctx context.Context, cfg *config.TestConfig) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
		cfg.DBName, cfg.DbSSLMode, cfg.DbMaxConn,
		cfg.DbMinConn, cfg.DbMaxConnLifeTime, cfg.DbMaxConnIdleTime,
	)

	return pkgpostgres.NewClient(ctx, pgConfig)
}

func setupLoad(t *testing.T) *testApp {
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

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn, // на нагрузке debug-логи только мешают
		}))

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

		// Seed dummy-пользователей (admin/user)
		err = userRepo.EnsureDummyUsers(ctx, cfg.DummyAdminID, cfg.DummyUserID)
		require.NoError(t, err)

		dummyLoginUC := usecase.NewDummyLoginUC(userRepo, jwtGen, cfg.DummyAdminID, cfg.DummyUserID)
		registerUC := usecase.NewRegisterUC(userRepo, passHasher)
		loginUC := usecase.NewLoginUC(userRepo, passHasher, jwtGen)
		createRoomUC := usecase.NewCreateRoomUC(roomRepo)
		listRoomsUC := usecase.NewListRoomsUC(roomRepo)
		createScheduleUC := usecase.NewCreateScheduleUC(trManager, roomRepo, scheduleRepo, slotRepo)
		listSlotsUC := usecase.NewListSlotsUC(trManager, roomRepo, scheduleRepo, slotRepo)
		createBookingUC := usecase.NewCreateBookingUC(trManager, slotRepo, bookingRepo, conferenceService)
		cancelBookingUC := usecase.NewCancelBookingUC(bookingRepo)
		listMyBookingsUC := usecase.NewListMyBookingsUC(bookingRepo)
		listBookingsUC := usecase.NewListBookingsUC(bookingRepo)

		authHandler := adapterhttp.NewAuthHandler(logger, dummyLoginUC, registerUC, loginUC)
		roomHandler := adapterhttp.NewRoomHandler(logger, createRoomUC, listRoomsUC)
		scheduleHandler := adapterhttp.NewScheduleHandler(logger, createScheduleUC)
		slotHandler := adapterhttp.NewSlotHandler(logger, listSlotsUC)
		bookingHandler := adapterhttp.NewBookingHandler(
			logger, createBookingUC, cancelBookingUC, listMyBookingsUC, listBookingsUC,
		)

		router := adapterhttp.NewRouter(
			authHandler, roomHandler, scheduleHandler, slotHandler, bookingHandler, jwtGen,
		).InitRoutes(logger)

		ts := httptest.NewServer(router)
		ts.Client().Transport = &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
		}

		appInstance = &testApp{
			server:   ts,
			client:   ts.Client(),
			pg:       pg,
			dbClient: pgClient,
			cfg:      cfg,
		}

		appInstance.seedData(t, ctx)
	})

	return appInstance
}

func (a *testApp) seedData(t *testing.T, ctx context.Context) {
	t.Log("seed: creating rooms + schedules...")
	a.roomIDs = make([]string, 0, seedRoomsCount)
	for i := 0; i < seedRoomsCount; i++ {
		roomID := a.createRoom(t, fmt.Sprintf("Room-%03d", i))
		a.createSchedule(t, roomID)
		a.roomIDs = append(a.roomIDs, roomID)
	}

	t.Log("seed: creating user pool...")
	a.userPool = make([]poolUser, 0, seedUsersCount)
	for i := 0; i < seedUsersCount; i++ {
		email := fmt.Sprintf("loaduser%d@test.local", i)
		pass := "load-test-pass-123"
		role := "user"
		id, token := a.createUser(t, &email, &pass, &role)
		a.userPool = append(a.userPool, poolUser{id: id, token: token})
	}

	t.Log("seed: collecting available slots across rooms/days...")
	// собираем слоты на ближайшие несколько дней по всем комнатам,
	// чтобы получить достаточный запас под исторические брони + под тест
	var allSlotIDs []string
	for dayOffset := 1; dayOffset <= 30 && len(allSlotIDs) < seedHistoricBookings+freshSlotsForBooking; dayOffset++ {
		date := time.Now().Add(time.Duration(dayOffset) * 24 * time.Hour).Format(time.DateOnly)
		for _, roomID := range a.roomIDs {
			slots := a.listSlotsRaw(t, roomID, date)
			for _, s := range slots {
				if id, ok := s["id"].(string); ok {
					allSlotIDs = append(allSlotIDs, id)
				}
			}
		}
	}
	require.NotEmpty(t, allSlotIDs, "no slots generated during seeding, check schedule config")

	historicCount := seedHistoricBookings
	if historicCount > len(allSlotIDs) {
		historicCount = len(allSlotIDs)
		t.Logf("seed: warning, only %d slots available for historic bookings (wanted %d)", historicCount, seedHistoricBookings)
	}

	t.Logf("seed: creating %d historic bookings (this may take a while)...", historicCount)
	a.bulkCreateBookings(t, allSlotIDs[:historicCount])

	// остаток слотов резервируем под сам нагрузочный прогон bookings/create
	remaining := allSlotIDs[historicCount:]
	a.freeSlots = make(chan string, len(remaining))
	for _, id := range remaining {
		a.freeSlots <- id
	}
	t.Logf("seed: reserved %d fresh slots for load run", len(a.freeSlots))
}

func (a *testApp) bulkCreateBookings(t *testing.T, slotIDs []string) {
	const seedWorkers = 50
	sem := make(chan struct{}, seedWorkers)
	var wg sync.WaitGroup

	for _, slotID := range slotIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(slotID string) {
			defer wg.Done()
			defer func() { <-sem }()

			user := a.userPool[rand.Intn(len(a.userPool))]
			payload := map[string]interface{}{"slot_id": slotID, "create_conference_link": false}
			resp, err := a.makeRequestAuth(http.MethodPost, "/bookings/create", payload, user.token)
			if err != nil {
				return
			}
			_ = resp.Body.Close()
		}(slotID)
	}
	wg.Wait()
}

func (a *testApp) cleanData(t *testing.T, ctx context.Context) {
	err := a.pg.TruncateTables(ctx)
	require.NoError(t, err, "failed to truncate tables")

	userRepo := adapterpostgres.NewUserRepository(a.dbClient, trmpgx.DefaultCtxGetter)
	err = userRepo.EnsureDummyUsers(ctx, a.cfg.DummyAdminID, a.cfg.DummyUserID)
	require.NoError(t, err, "failed to re-seed data")
}

// ──────────────────────────────────────────────────────────────
// HTTP helpers
// ──────────────────────────────────────────────────────────────

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
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, a.server.URL+path, &buf)
	if err != nil {
		return nil, err
	}
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

func (a *testApp) getUserToken(t *testing.T) string {
	if a.userToken != nil {
		return *a.userToken
	}
	token := a.getToken(t, "user")
	a.userToken = &token
	return token
}

func (a *testApp) getAdminToken(t *testing.T) string {
	if a.adminToken != nil {
		return *a.adminToken
	}
	token := a.getToken(t, "admin")
	a.adminToken = &token
	return token
}

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

	payload := map[string]interface{}{"email": userEmail, "password": userPass, "role": userRole}
	resp, err := a.makeRequest(http.MethodPost, registerPath, payload)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var user map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	id := user["user"]["id"].(string)
	assert.NotEmpty(t, id)

	payload = map[string]interface{}{"email": userEmail, "password": userPass}
	resp, err = a.makeRequest(http.MethodPost, loginPath, payload)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	require.NotEmpty(t, body["token"])

	token, ok := body["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	return id, token
}

func (a *testApp) createRoom(t *testing.T, name string) string {
	const path = "/rooms/create"
	payload := map[string]interface{}{"name": name, "capacity": 50}

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

func (a *testApp) createSchedule(t *testing.T, roomID string) {
	path := fmt.Sprintf("/rooms/%s/schedule/create", roomID)
	payload := map[string]interface{}{
		"days_of_week": []int{1, 2, 3, 4, 5, 6, 7},
		"start_time":   "08:00",
		"end_time":     "18:00",
	}

	resp, err := a.makeRequestAuth(http.MethodPost, path, payload, a.getAdminToken(t))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func (a *testApp) listSlotsRaw(t *testing.T, roomID, date string) []map[string]interface{} {
	path := fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, date)

	resp, err := a.makeRequestAuth(http.MethodGet, path, nil, a.getUserToken(t))
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var slots map[string][]map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		return nil
	}
	return slots["slots"]
}

// ──────────────────────────────────────────────────────────────
// Load runner
// ──────────────────────────────────────────────────────────────

type reqResult struct {
	elapsed time.Duration
	status  int
	err     error
}

func runLoad(rps int, workers int, duration time.Duration, fn func() (*http.Response, error)) []reqResult {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	limiter := rate.NewLimiter(rate.Limit(rps), rps)
	resultsCh := make(chan reqResult, rps*int(duration.Seconds()+1))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if err := limiter.Wait(ctx); err != nil {
					return
				}

				start := time.Now()
				resp, err := fn()
				elapsed := time.Since(start)

				status := 0
				if resp != nil {
					status = resp.StatusCode
					_ = resp.Body.Close()
				}

				select {
				case resultsCh <- reqResult{elapsed: elapsed, status: status, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]reqResult, 0, rps*int(duration.Seconds()+1))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func checkSLI(t *testing.T, name string, results []reqResult, latencyTarget time.Duration) {
	require.NotEmpty(t, results, "%s: no requests were made", name)

	var success int
	durations := make([]time.Duration, len(results))
	for i, r := range results {
		durations[i] = r.elapsed
		if r.err == nil && r.status >= 200 && r.status < 300 {
			success++
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	p50 := durations[len(durations)*50/100]
	p95 := durations[len(durations)*95/100]
	p99 := durations[len(durations)*99/100]
	successRate := float64(success) / float64(len(results)) * 100

	t.Logf(
		"[%s] total=%d success=%d success_rate=%.3f%% p50=%v p95=%v p99=%v",
		name, len(results), success, successRate, p50, p95, p99,
	)

	assert.GreaterOrEqualf(t, successRate, targetSuccessRate, "%s: success rate below SLI", name)
	if latencyTarget > 0 {
		assert.LessOrEqualf(t, p99, latencyTarget, "%s: p99 latency exceeds SLI", name)
	}
}

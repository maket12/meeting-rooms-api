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
	"strings"
	"sync"
	"testing"

	"backend/cmd/app/config"
	adapterhttp "backend/internal/adapter/in/http"
	adapterpg "backend/internal/adapter/out/postgres"
	adapteryacloud "backend/internal/adapter/out/yacloud"
	adapteryookassa "backend/internal/adapter/out/yookassa"
	"backend/internal/app/usecase"
	infrajwt "backend/internal/infrastructure/jwt"
	infrapass "backend/internal/infrastructure/password"
	infraqrcode "backend/internal/infrastructure/qrcode"
	"backend/internal/testhelpers"
	pkgpostgres "backend/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3Cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	migrationVersion = 7
	apiVersion       = "v1"
)

type testApp struct {
	server     *httptest.Server
	client     *http.Client
	pg         *testhelpers.PostgresContainer
	dbClient   *pkgpostgres.Client
	cfg        *config.Config
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

func newPostgresClient(ctx context.Context, cfg *config.Config) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
		cfg.DBName, cfg.DbSSLMode, cfg.DbMaxConn,
		cfg.DbMinConn, cfg.DbMaxConnLifeTime, cfg.DbMaxConnIdleTime,
	)

	return pkgpostgres.NewClient(ctx, pgConfig)
}

func newS3Client(ctx context.Context, internalCfg *config.Config) (*s3.Client, error) {
	cred := credentials.NewStaticCredentialsProvider(
		internalCfg.S3AccessKey,
		internalCfg.S3SecretKey,
		"",
	)

	cfg, err := s3Cfg.LoadDefaultConfig(
		ctx,
		s3Cfg.WithRegion(internalCfg.S3Region),
		s3Cfg.WithCredentialsProvider(cred),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 storage config: %w", err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(internalCfg.S3Endpoint)
	}), nil
}

func seedAdmins(
	ctx context.Context,
	cfg *config.Config,
	adminRepo *adapterpg.AdminRepository,
	passHasher *infrapass.Hasher,
) error {
	for _, seed := range cfg.GetAdminSeeds() {
		hash, err := passHasher.Hash(seed.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for login %s: %w", seed.Login, err)
		}

		// Поскольку тут UPSERT, метод безопасно отработает как на пустой, так и на заполненной базе
		err = adminRepo.EnsureAdmin(ctx, seed.Login, hash)
		if err != nil {
			return fmt.Errorf("failed to ensure admin for login %s: %w", seed.Login, err)
		}
	}
	return nil
}

func setupE2E(t *testing.T) *testApp {
	once.Do(func() {
		ctx := context.Background()

		_ = os.Setenv("DB_HOST", "test_host")
		_ = os.Setenv("DB_USER", "test_user")
		_ = os.Setenv("DB_PASSWORD", "test_pass")
		_ = os.Setenv("DB_NAME", "test_db")

		_ = os.Setenv("AUTH_SECRET", "super-secret-key-for-tests-32-chars!")
		_ = os.Setenv("AUTH_TTL", "24h")
		_ = os.Setenv("ADMIN_SEEDS", "test:test123")

		_ = os.Setenv("YOOKASSA_SHOP_ID", "1344511")
		_ = os.Setenv("YOOKASSA_API_KEY", "test_ivvXDKubIzQ-vo5_RMv5Z9a4zSQ9BHfhr7VybxhzabE")
		_ = os.Setenv("QR_CODE_BASE_URL", "http://localhost:8080")

		_ = os.Setenv("S3_BUCKET_NAME", "foodstock-test")
		_ = os.Setenv("S3_ACCESS_KEY", "YCAJEArxLfMxD5DYKG-b-6lSs")
		_ = os.Setenv("S3_SECRET_KEY", "YCOWxXBb6v21R6SDYw_wyn20iKLDg632t8jBxa6s")
		_ = os.Setenv("S3_ENDPOINT", "https://storage.yandexcloud.net")
		_ = os.Setenv("S3_REGION", "ru-central1")

		cfg, err := config.Load()
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

		adminRepo := adapterpg.NewAdminRepository(pgClient, trmpgx.DefaultCtxGetter)
		locationRepo := adapterpg.NewLocationRepository(pgClient, trmpgx.DefaultCtxGetter)
		itemRepo := adapterpg.NewItemRepository(pgClient, trmpgx.DefaultCtxGetter)
		locationItemRepo := adapterpg.NewLocationItemRepository(pgClient, trmpgx.DefaultCtxGetter)
		orderRepo := adapterpg.NewOrderRepository(pgClient, trmpgx.DefaultCtxGetter)
		orderItemRepo := adapterpg.NewOrderItemRepository(pgClient, trmpgx.DefaultCtxGetter)
		transactionRepo := adapterpg.NewTransactionRepository(pgClient, trmpgx.DefaultCtxGetter)

		tokenGen := infrajwt.NewGenerator(cfg.AuthSecret, cfg.AuthTTL)
		passHasher := infrapass.NewHasher(cfg.PasswordCost)
		qrCodeGen := infraqrcode.NewGenerator(cfg.QRCodeBaseURL, cfg.QRCodeSize)
		paymentGateway := adapteryookassa.NewPaymentGateway(cfg.YookassaShopID, cfg.YookassaAPIKey, cfg.YookassaTimeout)

		s3Client, err := newS3Client(ctx, cfg)
		require.NoError(t, err)
		mediaStorageGateway := adapteryacloud.NewYandexS3Storage(s3Client, cfg.S3BucketName)

		err = seedAdmins(ctx, cfg, adminRepo, passHasher)
		require.NoError(t, err)

		adminAuthUC := usecase.NewAdminAuthUC(adminRepo, passHasher, tokenGen)
		createLocationUC := usecase.NewCreateLocationUC(trManager, locationRepo, itemRepo, locationItemRepo)
		getLocationUC := usecase.NewGetLocationUC(locationRepo)
		updateLocationUC := usecase.NewUpdateLocationUC(locationRepo)
		deleteLocationUC := usecase.NewDeleteLocationUC(trManager, locationRepo, locationItemRepo)
		listLocationsUC := usecase.NewListLocationsUC(locationRepo)
		getQRCodeUC := usecase.NewGetQRCodeUC(locationRepo, qrCodeGen)
		getCatalogUC := usecase.NewGetCatalogUC(locationRepo, itemRepo, locationItemRepo)
		createItemUC := usecase.NewCreateItemUC(trManager, locationRepo, itemRepo, locationItemRepo)
		getItemUC := usecase.NewGetItemUC(itemRepo)
		updateItemUC := usecase.NewUpdateItemUC(itemRepo)
		deleteItemUC := usecase.NewDeleteItemUC(trManager, itemRepo, locationItemRepo)
		listItemsUC := usecase.NewListItemsUC(itemRepo)
		createOrderUC := usecase.NewCreateOrderUC(trManager, locationRepo, locationItemRepo, orderRepo, orderItemRepo, transactionRepo, paymentGateway)
		getInventoryUC := usecase.NewGetInventoryUC(locationRepo, locationItemRepo)
		updateInventoryUC := usecase.NewUpdateInventoryUC(trManager, locationRepo, locationItemRepo)
		getOrderStatusUC := usecase.NewGetOrderStatusUC(trManager, orderRepo, transactionRepo, paymentGateway)
		uploadMediaUC := usecase.NewUploadMediaUC(mediaStorageGateway)

		systemHandler := adapterhttp.NewSystemHandler(cfg.Environment, apiVersion)
		authHandler := adapterhttp.NewAuthHandler(logger, adminAuthUC)
		clientHandler := adapterhttp.NewClientHandler(logger, getCatalogUC, createOrderUC, getOrderStatusUC)
		locationsHandler := adapterhttp.NewLocationHandler(logger, createLocationUC, getLocationUC, updateLocationUC, deleteLocationUC, listLocationsUC, getQRCodeUC)
		itemHandler := adapterhttp.NewItemHandler(logger, createItemUC, getItemUC, updateItemUC, deleteItemUC, listItemsUC)
		inventoryHandler := adapterhttp.NewInventoryHandler(logger, getInventoryUC, updateInventoryUC)
		mediaHandler := adapterhttp.NewMediaHandler(logger, uploadMediaUC)

		router := adapterhttp.NewRouter(
			tokenGen, systemHandler,
			authHandler, clientHandler,
			locationsHandler, itemHandler,
			inventoryHandler, mediaHandler,
		).InitRoutes()

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
	tables := []string{
		"order_items",
		"orders",
		"transactions",
		"location_items",
		"items",
		"locations",
		"admins",
	}

	query := fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))

	_, err := a.dbClient.Pool.Exec(ctx, query)
	require.NoError(t, err, "failed to truncate tables")

	_, _ = a.dbClient.Pool.Exec(ctx, "DISCARD PLANS")

	adminRepo := adapterpg.NewAdminRepository(a.dbClient, trmpgx.DefaultCtxGetter)
	passHasher := infrapass.NewHasher(a.cfg.PasswordCost)
	err = seedAdmins(ctx, a.cfg, adminRepo, passHasher)
	require.NoError(t, err, "failed to re-seed admins")
}

func (a *testApp) doRequest(method, path string, body interface{}) (*http.Response, error) {
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

func (a *testApp) doRequestAuth(method, path string, body interface{}, token string) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req, _ := http.NewRequest(method, a.server.URL+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.client.Do(req)
}

func (a *testApp) getAdminToken(t *testing.T) string {
	if a.adminToken != nil {
		return *a.adminToken
	}

	resp, err := a.doRequest(
		"POST",
		"/api/v1/admin/auth",
		map[string]interface{}{
			"login":    "test",
			"password": "test123",
		},
	)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)

	require.NoError(t, err)
	require.NotEmpty(t, body["token"])

	token, ok := body["token"].(string)
	require.True(t, ok)

	a.adminToken = &token

	return token
}

// Helper for e2e tests.
// Creates the new location with specified parameters.  **
// Returns the slug of the created location.
// ** If a slug, name or an address are not specified, then it uses default values instead.
func (a *testApp) createLocation(t *testing.T, slug *string, name, address *string) string {
	const path = "/api/v1/admin/locations"

	var locSlug, locName, locAddress string

	if slug != nil {
		locSlug = *slug
	} else {
		locSlug = "test_1"
	}

	if name != nil {
		locName = *name
	} else {
		locName = "Test Location"
	}

	if address != nil {
		locAddress = *address
	} else {
		locAddress = "Address Of Test Location"
	}

	payload := map[string]interface{}{
		"slug":    locSlug,
		"name":    locName,
		"address": locAddress,
	}

	resp, err := a.doRequestAuth("POST", path, payload, a.getAdminToken(t))
	require.NoError(t, err)
	_ = resp.Body.Close()

	return locSlug
}

func (a *testApp) deleteLocation(t *testing.T, slug string) {
	path := fmt.Sprintf("/api/v1/admin/locations/%s", slug)
	resp, err := a.doRequestAuth("DELETE", path, nil, a.getAdminToken(t))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func (a *testApp) deactivateLocation(t *testing.T, slug string) {
	resp, err := a.doRequestAuth(
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

	resp, err := a.doRequestAuth("POST", path, payload, a.getAdminToken(t))
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
	resp, err := a.doRequestAuth(
		"DELETE",
		fmt.Sprintf("/api/v1/admin/items/%s", itemID),
		nil,
		a.getAdminToken(t),
	)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

// Helper for e2e tests.
// Sends some requests to prepare app for tests:
//  1. Create location with default values
//  2. Create several items with default values as well (price is 20050)
//  3. Update location items with default price and stock amount
//
// Returns: slug of location, ids of created items as a slice
func (a *testApp) seedInventoryData(t *testing.T, itemsCount int) (string, []string) {
	app := setupE2E(t)

	slug := app.createLocation(t, nil, nil, nil)

	itemIDs := make([]string, 0, itemsCount)
	for i := 0; i < itemsCount; i++ {
		itemIDs = append(itemIDs, app.createItem(t, nil))
	}

	inventory := make([]map[string]interface{}, 0, len(itemIDs))
	for _, id := range itemIDs {
		inventory = append(inventory, map[string]interface{}{
			"item_id":      id,
			"price":        20050,
			"is_available": true,
			"stock_amount": 5,
		})
	}

	path := fmt.Sprintf("/api/v1/admin/locations/%s/inventory", slug)
	body := map[string][]map[string]interface{}{"inventory": inventory}

	resp, err := a.doRequestAuth("PATCH", path, body, app.getAdminToken(t))
	require.NoError(t, err)
	_ = resp.Body.Close()

	return slug, itemIDs
}

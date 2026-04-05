package main

import (
	adapterhttp "MeetingRoomsAPI/internal/adapter/in/http"
	adapterpg "MeetingRoomsAPI/internal/adapter/out/postgres"
	"MeetingRoomsAPI/internal/app/usecase"
	"MeetingRoomsAPI/internal/config"
	infrahasher "MeetingRoomsAPI/internal/infrastructure/hasher"
	infrajwt "MeetingRoomsAPI/internal/infrastructure/jwt"
	pkgpostgres "MeetingRoomsAPI/pkg/postgres"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

const (
	shutdownTimeout = 10 * time.Second
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

	pgClient, err := pkgpostgres.NewClient(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	return pgClient, nil
}

func closePostgresClient(
	ctx context.Context,
	logger *slog.Logger,
	pgClient *pkgpostgres.Client,
) {
	logger.InfoContext(ctx, "closing postgres connection...")
	pgClient.Close()
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Postgres client
	pgClient, err := newPostgresClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres client: %w", err)
	}

	// Close Postgres
	defer closePostgresClient(ctx, logger, pgClient)

	// Transaction manager
	trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

	// Repositories
	userRepo := adapterpg.NewUserRepository(pgClient, trmpgx.DefaultCtxGetter)
	roomRepo := adapterpg.NewRoomRepository(pgClient, trmpgx.DefaultCtxGetter)
	scheduleRepo := adapterpg.NewScheduleRepository(pgClient, trmpgx.DefaultCtxGetter)
	slotRepo := adapterpg.NewSlotRepository(pgClient, trmpgx.DefaultCtxGetter)
	jwtGen := infrajwt.NewTokenGenerator(cfg.AuthSecret, cfg.AuthTTL)
	passHasher := infrahasher.NewPasswordHasher(cfg.PasswordCost)

	// Adding seed-data
	if err := userRepo.EnsureDummyUsers(ctx, cfg.DummyAdminID, cfg.DummyUserID); err != nil {
		return fmt.Errorf("failed to add seed-data: %w", err)
	}

	// Use-cases
	dummyLoginUC := usecase.NewDummyLoginUC(
		userRepo, jwtGen,
		cfg.DummyAdminID, cfg.DummyUserID,
	)
	registerUC := usecase.NewRegisterUC(userRepo, passHasher)
	loginUC := usecase.NewLoginUC(userRepo, passHasher, jwtGen)
	createRoomUC := usecase.NewCreateRoomUC(roomRepo)
	listRoomsUC := usecase.NewListRoomsUC(roomRepo)
	createScheduleUC := usecase.NewCreateScheduleUC(
		trManager, scheduleRepo, slotRepo,
	)
	listSlotsUC := usecase.NewListSlotsUC(trManager, slotRepo)

	// Handlers
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
	scheduleHandler := adapterhttp.NewScheduleHandler(
		logger,
		createScheduleUC,
	)
	slotHandler := adapterhttp.NewSlotHandler(
		logger,
		listSlotsUC,
	)

	router := adapterhttp.NewRouter(
		authHandler,
		roomHandler,
		scheduleHandler,
		slotHandler,
	).InitRoutes()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,

		BaseContext: func(_ net.Listener) context.Context { // any чтобы не тянуть net.Listener в импорты
			return ctx
		},
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("starting server", slog.String("address", ":8080"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("server failed", slog.Any("err", err))
			return err
		}
		logger.Info("server stopped")
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
		_ = srv.Close() // fallback
		return err
	}

	logger.Info("server exited properly")
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runServer(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}

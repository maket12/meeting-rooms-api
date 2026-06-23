package main

import (
	"backend/cmd/app/config"
	http2 "backend/internal/adapter/in/http"
	adapterconf "backend/internal/adapter/out/conference"
	"backend/internal/adapter/out/postgres"
	usecase2 "backend/internal/app/usecase"
	infrajwt "backend/internal/infrastructure/jwt"
	infrapasswd "backend/internal/infrastructure/password"
	pkgpostgres "backend/pkg/postgres"
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
	userRepo := postgres.NewUserRepository(pgClient, trmpgx.DefaultCtxGetter)
	roomRepo := postgres.NewRoomRepository(pgClient, trmpgx.DefaultCtxGetter)
	scheduleRepo := postgres.NewScheduleRepository(pgClient, trmpgx.DefaultCtxGetter)
	slotRepo := postgres.NewSlotRepository(pgClient, trmpgx.DefaultCtxGetter)
	bookingRepo := postgres.NewBookingRepository(pgClient, trmpgx.DefaultCtxGetter)
	conferenceService := adapterconf.NewConferenceService("available")
	jwtGen := infrajwt.NewTokenGenerator(cfg.AuthSecret, cfg.AuthTTL)
	passHasher := infrapasswd.NewHasher(cfg.PasswordCost)

	// Adding seed-data
	if err = userRepo.EnsureDummyUsers(ctx, cfg.DummyAdminID, cfg.DummyUserID); err != nil {
		return fmt.Errorf("failed to add seed-data: %w", err)
	}

	// Use-cases
	dummyLoginUC := usecase2.NewDummyLoginUC(
		userRepo, jwtGen,
		cfg.DummyAdminID, cfg.DummyUserID,
	)
	registerUC := usecase2.NewRegisterUC(userRepo, passHasher)
	loginUC := usecase2.NewLoginUC(userRepo, passHasher, jwtGen)
	createRoomUC := usecase2.NewCreateRoomUC(roomRepo)
	listRoomsUC := usecase2.NewListRoomsUC(roomRepo)
	createScheduleUC := usecase2.NewCreateScheduleUC(
		trManager, roomRepo, scheduleRepo, slotRepo,
	)
	listSlotsUC := usecase2.NewListSlotsUC(
		trManager, roomRepo, scheduleRepo, slotRepo,
	)
	createBookingUC := usecase2.NewCreateBookingUC(
		trManager, slotRepo,
		bookingRepo, conferenceService,
	)
	cancelBookingUC := usecase2.NewCancelBookingUC(bookingRepo)
	listMyBookingsUC := usecase2.NewListMyBookingsUC(bookingRepo)
	listBookingsUC := usecase2.NewListBookingsUC(bookingRepo)

	// Handlers
	authHandler := http2.NewAuthHandler(
		logger,
		dummyLoginUC,
		registerUC,
		loginUC,
	)
	roomHandler := http2.NewRoomHandler(
		logger,
		createRoomUC,
		listRoomsUC,
	)
	scheduleHandler := http2.NewScheduleHandler(
		logger, createScheduleUC,
	)
	slotHandler := http2.NewSlotHandler(logger, listSlotsUC)
	bookingHandler := http2.NewBookingHandler(
		logger,
		createBookingUC,
		cancelBookingUC,
		listMyBookingsUC,
		listBookingsUC,
	)

	router := http2.NewRouter(
		authHandler,
		roomHandler,
		scheduleHandler,
		slotHandler,
		bookingHandler,
		jwtGen,
	).InitRoutes(logger)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,

		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)

	go func() {
		logger.InfoContext(ctx, "starting server", slog.String("address", ":8080"))
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.InfoContext(ctx, "shutdown signal received")
	case err = <-errCh:
		if err != nil {
			logger.ErrorContext(ctx, "server failed", slog.Any("err", err))
			return err
		}
		logger.InfoContext(ctx, "server stopped")
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.ErrorContext(ctx, "graceful shutdown failed", slog.Any("err", err))
		_ = srv.Close() // fallback
		return err
	}

	logger.InfoContext(ctx, "server exited properly")
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

	if err = runServer(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}

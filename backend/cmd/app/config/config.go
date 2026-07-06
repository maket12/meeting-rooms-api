package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
)

type Config struct {
	// Postgres
	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"DB_USER" envDefault:"meeting-rooms-user"`
	DbPassword string `env:"DB_PASSWORD" envDefault:"meeting-rooms-password"`
	DBName     string `env:"DB_NAME" envDefault:"meeting-rooms-db"`
	DbSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	DbMaxConn         int           `env:"DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Auth constants
	AuthSecret   string        `env:"AUTH_SECRET" envDefault:"super-secret-token"`
	AuthTTL      time.Duration `env:"AUTH_TTL" envDefault:"1h"`
	DummyAdminID uuid.UUID     `env:"DUMMY_ADMIN_ID" envDefault:"00000000-0000-0000-0000-000000000001"`
	DummyUserID  uuid.UUID     `env:"DUMMY_USER_ID" envDefault:"00000000-0000-0000-0000-000000000002"`

	// Password hasher
	PasswordCost int `env:"PASSWORD_COST" envDefault:"10"`

	// Service
	HttpPort    int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.DbHost)
	fmt.Printf("   HTTP Port: %d\n", cfg.HttpPort)

	return cfg, nil
}

type TestConfig struct {
	// Postgres
	DbHost     string `env:"TEST_DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"TEST_DB_PORT" envDefault:"5433"`
	DbUser     string `env:"TEST_DB_USER" envDefault:"test"`
	DbPassword string `env:"TEST_DB_PASSWORD" envDefault:"test123"`
	DBName     string `env:"TEST_DB_NAME" envDefault:"test-db"`
	DbSSLMode  string `env:"TEST_DB_SSL_MODE" envDefault:"disable"`

	DbMaxConn         int           `env:"TEST_DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"TEST_DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"TEST_DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"TEST_DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Auth constants
	AuthSecret   string        `env:"TEST_AUTH_SECRET" envDefault:"test-token"`
	AuthTTL      time.Duration `env:"TEST_AUTH_TTL" envDefault:"1h"`
	DummyAdminID uuid.UUID     `env:"TEST_DUMMY_ADMIN_ID" envDefault:"00000000-0000-0000-0000-000000000001"`
	DummyUserID  uuid.UUID     `env:"TEST_DUMMY_USER_ID" envDefault:"00000000-0000-0000-0000-000000000002"`

	// Password hasher
	PasswordCost int `env:"TEST_PASSWORD_COST" envDefault:"4"`

	// Service
	HttpPort    int    `env:"TEST_HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"TEST_LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"TEST_ENVIRONMENT" envDefault:"test"`
}

func LoadTest() (*TestConfig, error) {
	cfg := &TestConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load test config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.DbHost)
	fmt.Printf("   HTTP Port: %d\n", cfg.HttpPort)

	return cfg, nil
}

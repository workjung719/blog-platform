package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB      DBConfig
	JWT     JWTConfig
	Server  ServerConfig
	Sched   SchedulerConfig
	Logging LogConfig
}

type DBConfig struct {
	Host, Port, User, Password, Name, SSLMode string
}

func (c DBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

type JWTConfig struct {
	Secret          []byte
	ExpirationHours int
}

type ServerConfig struct{ Port string }

type SchedulerConfig struct {
	IntervalSec int
	Workers     int
}

type LogConfig struct{ Level string }

func Load() (*Config, error) {
	_ = godotenv.Load() // отсутствие .env не фатально — есть дефолты

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5436"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "blog_platform"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          []byte(secret),
			ExpirationHours: getInt("JWT_EXPIRATION_HOURS", 24),
		},
		Server: ServerConfig{Port: getEnv("SERVER_PORT", "8080")},
		Sched: SchedulerConfig{
			IntervalSec: getInt("SCHEDULER_INTERVAL_SEC", 5),
			Workers:     getInt("SCHEDULER_WORKERS", 3),
		},
		Logging: LogConfig{Level: getEnv("LOG_LEVEL", "info")},
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
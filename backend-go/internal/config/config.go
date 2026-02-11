package config

import (
	"os"
)

type Config struct {
	GoPort    string
	WsPort    string
	DBURL     string
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	return Config{
		GoPort: getenv("GO_PORT", "8080"),
		WsPort: getenv("GO_WS_PORT", "8081"),
		DBURL:  getenv("DATABASE_URL", "postgres://autotrade:autotrade@localhost:5432/autotrade?sslmode=disable"),
	}
}

package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	App      `mapstructure:",squash"`
	Database `mapstructure:",squash"`
	Redis    `mapstructure:",squash"`
	Logger   `mapstructure:",squash"`
}

type App struct {
	// Port is injected by most container hosts (Render, Fly, Koyeb).
	Port string `env:"PORT" env-default:"8080"`
	// AllowedOrigins is a comma-separated CORS allowlist. Only needed when the
	// frontend is served from somewhere other than this binary (e.g. GitHub Pages).
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" env-separator:","`
	// CreateRateLimit is the sustained link-creation rate per client IP, in
	// requests per second; CreateRateBurst is the short spike allowed above it.
	CreateRateLimit float64 `env:"CREATE_RATE_LIMIT" env-default:"5"`
	CreateRateBurst int     `env:"CREATE_RATE_BURST" env-default:"10"`
}

type Database struct {
	DBHost   string `env:"DBHOST"`
	DbUser   string `env:"DBUSER"`
	DbPass   string `env:"DBPASS"`
	DbPort   string `env:"DBPORT"`
	DbName   string `env:"DBNAME"`
	DbSchema string `env:"DBSCHEMA"`
	// DbSSLMode must be "require" (or stricter) for managed Postgres such as Neon.
	DbSSLMode string `env:"DBSSLMODE" env-default:"disable"`
	Debug     bool   `env:"DEBUG" env-default:"false"`
}

type Redis struct {
	RedisAddress  string `env:"REDIS_ADDRESS"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	// RedisTLS must be true for managed Redis such as Upstash.
	RedisTLS bool `env:"REDIS_TLS" env-default:"false"`
}

type Logger struct {
	Development       bool   `env:"LOG_DEVELOPMENT"`
	DisableCaller     bool   `env:"LOG_DISABLE_CALLER"`
	DisableStacktrace bool   `env:"LOG_DISABLE_STACKTRACE"`
	Encoding          string `env:"LOG_ENCODING"`
	Level             string `env:"LOG_LEVEL"`
}

func NewConfig(configFile string) *Config {
	config := Config{}
	if err := godotenv.Load(configFile); err != nil {
		log.Printf("no %s file found, falling back to process environment: %v", configFile, err)
	}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatalln(err)
	}
	return &config
}

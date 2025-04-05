package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Database `mapstructure:",squash"`
	Redis    `mapstructure:",squash"`
	Logger   `mapstructure:",squash"`
}

type Database struct {
	DBHost   string `env:"DBHOST"`
	DbUser   string `env:"DBUSER"`
	DbPass   string `env:"DBPASS"`
	DbPort   string `env:"DBPORT"`
	DbName   string `env:"DBNAME"`
	DbSchema string `env:"DBSCHEMA"`
	Debug    bool   `env:"DEBUG" env-default:"false"`
}

type Redis struct {
	RedisAddress string `env:"REDIS_ADDRESS"`
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
	err := godotenv.Load(configFile)
	if err != nil {
		return nil
	}
	err = cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatalln(err)
	}
	return &config
}

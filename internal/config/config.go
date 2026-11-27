package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	authClient "gitlab.com/mlplabs/auth-service/pkg/client"
)

type (
	Config struct {
		HTTP `envPrefix:"HTTP_"`
		Log
		PgDb       `envPrefix:"DB_"`
		Auth       `envPrefix:"AUTH_"`
		AuthClient authClient.Config `envPrefix:"AUTH_CLIENT_"`
	}
	HTTP struct {
		Host string `env:"HOST" envDefault:"localhost"`
		Port string `env:"PORT" envDefault:"8080"`
	}
	PgDb struct {
		PoolMax  int    `env:"POOL_MAX"             envDefault:"3"`
		Host     string `env:"HOST,required"`
		Port     int    `env:"PORT,required"`
		User     string `env:"USERNAME,required"`
		Password string `env:"PASSWORD,required"`
		DB       string `env:"NAME,required"`
	}
	Auth struct {
		Mode           string `env:"MODE"              envDefault:"none"` // local, external
		JwtSecret      string `env:"JWT_SECRET"        envDefault:""`     // в base64 для local или внешнего сервиса с общим секретом (HS256)
		JwrKeysBaseUrl string `env:"JWT_KEYS_BASE_URL" envDefault:""`     // для внешнего сервиса с публичными ключами (RS256/ES256), если JwtSecret пустой
	}

	Log struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}
)

func NewConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	return &cfg, nil
}

func (cfg *Config) IsAuthDisable() bool {
	return cfg.Auth.Mode == "" || cfg.Auth.Mode == "none"
}

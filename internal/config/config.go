package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppPort        string `env:"APP_PORT" envDefault:"8080"`
	AppEnv         string `env:"APP_ENV" envDefault:"development"`
	DBHost         string `env:"DB_HOST,required"`
	DBPort         int    `env:"DB_PORT" envDefault:"5432"`
	DBUser         string `env:"DB_USER,required"`
	DBPassword     string `env:"DB_PASSWORD,required"`
	DBName         string `env:"DB_NAME,required"`
	DBSSLMode      string `env:"DB_SSL_MODE" envDefault:"disable"`
	JWTSecret      string `env:"JWT_SECRET,required"`
	JWTExpiryHours int    `env:"JWT_EXPIRY_HOURS" envDefault:"72"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	return cfg, env.Parse(cfg)
}

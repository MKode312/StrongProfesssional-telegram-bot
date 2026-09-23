package config

import (
	"errors"
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"APP_ENV" env-default:"local"`
	Telegram TelegramConfig `yaml:"telegram"`
	Postgres PostgresConfig `yaml:"postgres"`
	SMTP     SMTPConfig     `yaml:"smtp"`
}

type TelegramConfig struct {
	Token string `yaml:"token" env:"BOT_TOKEN" env-required:"true"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     uint16 `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	Name     string `yaml:"name" env:"POSTGRES_DB" env-required:"true"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
}

type SMTPConfig struct {
	Host         string `yaml:"host" env:"SMTP_HOST" env-required:"true"`
	Port         int    `yaml:"port" env:"SMTP_PORT" env-default:"587"`
	Username     string `yaml:"username" env:"SMTP_USERNAME" env-required:"true"`
	Password     string `yaml:"password" env:"SMTP_PASSWORD" env-required:"true"`
	From         string `yaml:"from" env:"SMTP_FROM" env-required:"true"`
	CompanyEmail string `yaml:"company_email" env:"COMPANY_EMAIL" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic("load .env: " + err.Error())
	}

	path := "config/local.yaml"
	flag.StringVar(&path, "config", path, "path to YAML configuration")
	flag.Parse()

	if _, err := os.Stat(path); err != nil {
		panic("config file: " + err.Error())
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("read configuration: " + err.Error())
	}
	return &cfg
}

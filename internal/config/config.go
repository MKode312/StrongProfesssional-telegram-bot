package config

import (
	"errors"
	"flag"
	"net"
	"net/url"
	"os"
	"strconv"

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
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"postgres"`
	Port     uint16 `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	Name     string `yaml:"name" env:"POSTGRES_DB" env-required:"true"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
}

func (c PostgresConfig) URL() string {
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.FormatUint(uint64(c.Port), 10)),
		Path:   c.Name,
	}
	query := dsn.Query()
	query.Set("sslmode", c.SSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
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
	var cfg Config
	mustLoad(&cfg)

	return &cfg
}

func MustLoadPostgres() PostgresConfig {
	var cfg struct {
		Postgres PostgresConfig `yaml:"postgres"`
	}
	mustLoad(&cfg)

	return cfg.Postgres
}

func mustLoad(cfg any) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic("load .env: " + err.Error())
	}

	path := "config/local.yaml"
	flag.StringVar(&path, "config", path, "path to YAML configuration")
	flag.Parse()

	if _, err := os.Stat(path); err != nil {
		panic("config file: " + err.Error())
	}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic("read configuration: " + err.Error())
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"str-prof-bot/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

func main() {
	migrationsPath, ok := os.LookupEnv("MIGRATIONS_PATH")
	if !ok {
		panic("migrations-path is required")
	}

	postgresConfig := config.MustLoadPostgres()
	databaseURL := postgresConfig.URL()

	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"pet-link/cmd/migrator/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	up   = flag.Bool("up", false, "apply all up migrations")
	down = flag.Int("down", 0, "rollback N migrations")
)

func main() {
	flag.Parse()
	if !*up && *down == 0 {
		flag.PrintDefaults()
		return
	}

	dbConnectionConfig := config.LoadDatabaseConfig()
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConnectionConfig.User,
		dbConnectionConfig.Password,
		dbConnectionConfig.Host,
		dbConnectionConfig.Port,
		dbConnectionConfig.Name,
	)

	migrationTable := "migrations"
	migrationsPath := "file://migrations"
	dbURLWithTable := fmt.Sprintf("%s&x-migrations-table=%s", dbURL, migrationTable)

	if err := runMigrations(migrationsPath, dbURLWithTable, *up, *down); err != nil {
		log.Fatalf("migrator error: %v", err)
	}
}

func runMigrations(migrationsPath, dbURLWithTable string, up bool, down int) error {
	m, err := migrate.New(migrationsPath, dbURLWithTable)
	if err != nil {
		return fmt.Errorf("error init migrations: %w", err)
	}
	defer m.Close()

	switch {
	case up:
		err := m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no new migrations to apply")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println("migrations applied")
		return nil

	case down > 0:
		if err := m.Steps(-down); err != nil {
			return fmt.Errorf("error rolling back migrations: %w", err)
		}
		fmt.Printf("rolling back %d migrations\n", down)
		return nil

	default:
		return nil
	}
}

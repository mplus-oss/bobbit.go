package metadata

import (
	"embed"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/internal/dblib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func InitDB(cfg config.BobbitDaemonConfig) (*sqlx.DB, error) {
	log.Println("Connecting to local database.")
	db, err := sqlx.Open(
		"sqlite3",
		path.Join(cfg.DataPath, "metadata.db"),
	)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConn)
	db.SetMaxIdleConns(cfg.DBMaxIdleConn)
	db.SetConnMaxLifetime(time.Hour)

	log.Println("Running the migration")
	err = runMigration(db)
	if err != nil {
		return nil, fmt.Errorf("Failed to migrate: %w", err)
	}

	// Check SQLite version for JSON function support
	if _, err := dblib.CheckSQLiteJSONFunctions(db); err != nil {
		log.Printf("[WARNING] Failed to check SQLite JSON function support: %v", err)
	}

	return db, nil
}

func runMigration(db *sqlx.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("Failed to create migration source: %w", err)
	}

	dbDriver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("Failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"sqlite3", dbDriver,
	)
	if err != nil {
		return fmt.Errorf("Failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	if err == migrate.ErrNoChange {
		log.Println("Database schema is up to date.")
	} else {
		log.Println("Database migrations applied successfully.")
	}

	enablePragma(db, map[string]string{
		"journal_mode": "WAL",
		"foreign_keys": "ON",
		"busy_timeout": "5000",
		"synchronous":  "NORMAL",
	})

	return nil
}

func enablePragma(db *sqlx.DB, pragmas map[string]string) {
	for k, v := range pragmas {
		query := fmt.Sprintf("PRAGMA %s = %s;", k, v)
		if _, err := db.Exec(query); err != nil {
			log.Printf("[WARNING] Cannot enable pragma %v: %v\n", query, err)
		}
	}
	log.Println("Database Pragma enabled.")
}

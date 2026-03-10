package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type SSLMode string

var Assets *sqlx.DB

const (
	SSLModeDisabled SSLMode = "disable"
	SSLModeEnabled  SSLMode = "enable"
)

//connect -ping-save globlly-run migrations

func CreateAndMigrate(host, port, user, password, dbname string, sslmode SSLMode) error {
	connStr := fmt.Sprintf("host =%s port = %s user =%s password =%s dbname =%s sslmode=%s ", host, port, user, password, dbname, sslmode)

	log.Println("DB host:", host)
	log.Println("DB port:", port)
	log.Println("DB user:", user)
	log.Println("DB name:", dbname)

	DataBase, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	err = DataBase.Ping() //testing
	if err != nil {
		return err
	}
	Assets = DataBase
	return migrateUp(DataBase)
}

func migrateUp(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Print("Migration successful")

	return nil
}

func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := Assets.Beginx()
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}

	defer func() {

		if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Printf("could not rollback transaction: %v", rollBackErr)
			}
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("Could not commit transaction: %v", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

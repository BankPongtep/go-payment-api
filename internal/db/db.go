package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sqlx.DB
}

func Connect(dsn string) (*DB, error) {
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/payments_db?sslmode=disable"
	}

	var db *sqlx.DB
	var err error

	for i := 1; i <= 3; i++ {
		db, err = sqlx.Connect("pgx", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		fmt.Printf("Attempt %d: failed to connect to DB: %v. Retrying in 2s...\n", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("connect db after 3 attempts: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

package main

import (
	"os"

	"github.com/BankPongtep/go-payment-api/internal/db"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	dsn := os.Getenv("DATABASE_URL")
	database, err := db.Connect(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	log.Info().Msg("Migration finished")
}

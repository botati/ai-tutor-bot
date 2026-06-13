package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/cobrich/ai-tutor-bot/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()

	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}

	db, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("failed to set goose dialect:", err)
	}

	command := "up"

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if err := goose.Run(command, db, "internal/migrations/sql"); err != nil {
		log.Fatal("failed to run migrations:", err)
	}

	log.Println("migrations completed:", command)
}

package main

import (
	"activity-bot/internal/config"
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed", err)
	}

	dsn := cfg.DBDSN
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}

	if dsn == "" {
		log.Fatal("DB_DSN is required (via .env or environment variable)")
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Usage: migrate [status|up|down|redo|...]")
	}

	command := args[0]
	cmdArgs := args[1:]

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.RunContext(ctx, command, db, "migrations", cmdArgs...); err != nil {
		log.Fatalf("goose run failed: %v", err)
	}
}

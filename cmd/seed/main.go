package main

import (
	"activity-bot/internal/config"
	db "activity-bot/internal/db/postgres/sqlc"

	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const wordsURL = "https://raw.githubusercontent.com/LussRus/Rus_words/refs/heads/master/UTF8/txt/nouns/summary.txt"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.New(
		ctx,
		cfg.DBDSN,
	)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	queries := db.New(pool)

	if err := seedWords(ctx, queries); err != nil {
		panic(err)
	}
}

func seedWords(
	ctx context.Context,
	queries *db.Queries,
) error {
	resp, err := http.Get(wordsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"download failed: %s",
			resp.Status,
		)
	}

	scanner := bufio.NewScanner(resp.Body)

	scanner.Buffer(
		make([]byte, 1024),
		1024*1024,
	)

	count := 0

	for scanner.Scan() {
		word := normalize(scanner.Text())

		if !validWord(word) {
			continue
		}

		err := queries.InsertCrocodileWord(
			ctx,
			word,
		)

		if err != nil {
			return err
		}

		count++

		if count%1000 == 0 {
			fmt.Printf(
				"inserted %d words\n",
				count,
			)
		}
	}

	fmt.Printf(
		"done: %d words\n",
		count,
	)

	return scanner.Err()
}

func normalize(s string) string {
	return strings.ToLower(
		strings.TrimSpace(s),
	)
}

func validWord(s string) bool {
	if len([]rune(s)) < 3 {
		return false
	}

	for _, r := range s {
		if r < 'а' || r > 'я' {
			if r != 'ё' {
				return false
			}
		}
	}

	return true
}

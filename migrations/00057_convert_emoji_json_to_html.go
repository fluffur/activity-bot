package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(UP00057, DOWN00057)
}

func UP00057(ctx context.Context, tx *sql.Tx) error {
	var cmRows []struct {
		chatID int64
		userID int64
		raw    string
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT chat_id, user_id, emoji_json
		FROM chat_members
		WHERE emoji_json IS NOT NULL
	`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var r struct {
			chatID int64
			userID int64
			raw    string
		}

		if err := rows.Scan(&r.chatID, &r.userID, &r.raw); err != nil {
			_ = rows.Close()

			return err
		}

		cmRows = append(cmRows, r)
	}

	_ = rows.Close()

	for _, r := range cmRows {
		var parsed []Emoji

		if err := json.Unmarshal([]byte(r.raw), &parsed); err != nil {
			return err
		}

		var result strings.Builder

		for _, e := range parsed {
			if e.Type == TypeCustom {
				result.WriteString(fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, e.ID, e.Char))
			} else {
				result.WriteString(e.Char)
			}
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE chat_members
			SET emoji = $1
			WHERE chat_id = $2 AND user_id = $3
		`, result.String(), r.chatID, r.userID); err != nil {
			return err
		}
	}

	var uRows []struct {
		id  int64
		raw string
	}

	rows2, err := tx.QueryContext(ctx, `
		SELECT id, emoji_json
		FROM users
		WHERE emoji_json IS NOT NULL
	`)
	if err != nil {
		return err
	}

	for rows2.Next() {
		var r struct {
			id  int64
			raw string
		}

		if err := rows2.Scan(&r.id, &r.raw); err != nil {
			_ = rows2.Close()

			return err
		}

		uRows = append(uRows, r)
	}

	_ = rows2.Close()

	for _, r := range uRows {
		var parsed []Emoji

		if err := json.Unmarshal([]byte(r.raw), &parsed); err != nil {
			return err
		}

		var result strings.Builder

		for _, e := range parsed {
			if e.Type == TypeCustom {
				result.WriteString(fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, e.ID, e.Char))
			} else {
				result.WriteString(e.Char)
			}
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET emoji = $1
			WHERE id = $2
		`, result.String(), r.id); err != nil {
			return err
		}
	}

	return nil
}

func DOWN00057(_ context.Context, _ *sql.Tx) error {
	return nil
}

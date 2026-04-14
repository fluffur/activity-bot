package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/rivo/uniseg"
)

func init() {
	goose.AddMigrationContext(UP00048, DOWN00048)
}

type Emoji struct {
	Type string `json:"type"`
	ID   int64  `json:"id,omitempty"`
	Char string `json:"char"`
}

var tgEmojiFullRegex = regexp.MustCompile(`<tg-emoji emoji-id="(\d+)">(.+?)</tg-emoji>`)
var tgEmojiStartRegex = regexp.MustCompile(`^<tg-emoji`)

func parseMixed(input string) []Emoji {
	var result []Emoji

	for len(input) > 0 {
		if tgEmojiStartRegex.MatchString(input) {
			loc := tgEmojiFullRegex.FindStringIndex(input)
			if loc != nil && loc[0] == 0 {
				tag := input[:loc[1]]
				sub := tgEmojiFullRegex.FindStringSubmatch(tag)

				if len(sub) == 3 {
					var id int64
					for _, c := range sub[1] {
						id = id*10 + int64(c-'0')
					}

					result = append(result, Emoji{
						Type: "custom",
						ID:   id,
						Char: sub[2],
					})

					input = input[loc[1]:]
					continue
				}
			}
		}

		g := uniseg.NewGraphemes(input)
		if g.Next() {
			part := g.Str()

			result = append(result, Emoji{
				Type: "unicode",
				Char: part,
			})

			input = input[len(part):]
			continue
		}

		break
	}

	return result
}

type cmRow struct {
	chatID int64
	userID int64
	raw    string
}

type uRow struct {
	id  int64
	raw string
}

func UP00048(ctx context.Context, tx *sql.Tx) error {
	var cmRows []cmRow

	rows, err := tx.QueryContext(ctx, `
		SELECT chat_id, user_id, emoji
		FROM chat_members
		WHERE emoji IS NOT NULL AND emoji != ''
	`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var r cmRow
		if err := rows.Scan(&r.chatID, &r.userID, &r.raw); err != nil {
			rows.Close()
			return err
		}
		cmRows = append(cmRows, r)
	}
	rows.Close()

	for _, r := range cmRows {
		parsed := parseMixed(r.raw)
		if len(parsed) == 0 {
			continue
		}

		b, err := json.Marshal(parsed)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE chat_members
			SET emoji_json = $1
			WHERE chat_id = $2 AND user_id = $3
		`, b, r.chatID, r.userID); err != nil {
			return err
		}
	}

	var uRows []uRow

	rows2, err := tx.QueryContext(ctx, `
		SELECT id, emoji
		FROM users
		WHERE emoji IS NOT NULL AND emoji != ''
	`)
	if err != nil {
		return err
	}

	for rows2.Next() {
		var r uRow
		if err := rows2.Scan(&r.id, &r.raw); err != nil {
			rows2.Close()
			return err
		}
		uRows = append(uRows, r)
	}
	rows2.Close()

	for _, r := range uRows {
		parsed := parseMixed(r.raw)
		if len(parsed) == 0 {
			continue
		}

		b, err := json.Marshal(parsed)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET emoji_json = $1
			WHERE id = $2
		`, b, r.id); err != nil {
			return err
		}
	}

	return nil
}

func DOWN00048(ctx context.Context, tx *sql.Tx) error {
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
			rows.Close()
			return err
		}
		cmRows = append(cmRows, r)
	}
	rows.Close()

	for _, r := range cmRows {
		var parsed []Emoji
		if err := json.Unmarshal([]byte(r.raw), &parsed); err != nil {
			return err
		}

		var result strings.Builder

		for _, e := range parsed {
			if e.Type == "custom" {
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
			rows2.Close()
			return err
		}
		uRows = append(uRows, r)
	}
	rows2.Close()

	for _, r := range uRows {
		var parsed []Emoji
		if err := json.Unmarshal([]byte(r.raw), &parsed); err != nil {
			return err
		}

		var result strings.Builder

		for _, e := range parsed {
			if e.Type == "custom" {
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

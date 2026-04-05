package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(UP00048, DOWN00048)
}

type Emoji struct {
	ID   int64  `json:"id"`
	Char string `json:"char"`
}

var tgEmojiRegex = regexp.MustCompile(`<tg-emoji emoji-id="(\d+)">(.+?)</tg-emoji>`)

func parseEmojiHTML(raw string) []Emoji {
	var emojis []Emoji
	matches := tgEmojiRegex.FindAllStringSubmatch(raw, -1)
	for _, m := range matches {
		var id int64
		for _, c := range m[1] {
			id = id*10 + int64(c-'0')
		}
		emojis = append(emojis, Emoji{
			ID:   id,
			Char: m[2],
		})
	}
	return emojis
}

func UP00048(ctx context.Context, tx *sql.Tx) error {
	// 1. Process chat_members
	type cmTask struct {
		chatID, userID int64
		emojiJSON      string
	}
	var cmTasks []cmTask

	rows, err := tx.QueryContext(ctx, `
		SELECT chat_id, user_id, emoji 
		FROM chat_members 
		WHERE emoji IS NOT NULL AND emoji != ''
	`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var chatID, userID int64
		var emojiHTML string

		if err := rows.Scan(&chatID, &userID, &emojiHTML); err != nil {
			rows.Close()
			return err
		}

		emojis := parseEmojiHTML(emojiHTML)
		if len(emojis) > 0 {
			b, err := json.Marshal(emojis)
			if err != nil {
				rows.Close()
				return err
			}
			cmTasks = append(cmTasks, cmTask{chatID, userID, string(b)})
		}
	}
	rows.Close()

	for _, t := range cmTasks {
		if _, err := tx.ExecContext(ctx, `
			UPDATE chat_members 
			SET emoji_json = $1 
			WHERE chat_id = $2 AND user_id = $3
		`, t.emojiJSON, t.chatID, t.userID); err != nil {
			return err
		}
	}

	// 2. Process users
	type uTask struct {
		id        int64
		emojiJSON string
	}
	var uTasks []uTask

	rows2, err := tx.QueryContext(ctx, `
		SELECT id, emoji 
		FROM users 
		WHERE emoji IS NOT NULL AND emoji != ''
	`)
	if err != nil {
		return err
	}

	for rows2.Next() {
		var id int64
		var emojiHTML string

		if err := rows2.Scan(&id, &emojiHTML); err != nil {
			rows2.Close()
			return err
		}

		emojis := parseEmojiHTML(emojiHTML)
		if len(emojis) > 0 {
			b, err := json.Marshal(emojis)
			if err != nil {
				rows2.Close()
				return err
			}
			uTasks = append(uTasks, uTask{id, string(b)})
		}
	}
	rows2.Close()

	for _, t := range uTasks {
		if _, err := tx.ExecContext(ctx, `
			UPDATE users 
			SET emoji_json = $1 
			WHERE id = $2
		`, t.emojiJSON, t.id); err != nil {
			return err
		}
	}

	return nil
}

func DOWN00048(ctx context.Context, tx *sql.Tx) error {
	type cmDownTask struct {
		chatID, userID int64
		emojiHTML      string
	}
	var cmDownTasks []cmDownTask

	rows, err := tx.QueryContext(ctx, `SELECT chat_id, user_id, emoji_json FROM chat_members WHERE emoji_json IS NOT NULL`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var chatID, userID int64
		var jsonStr string
		if err := rows.Scan(&chatID, &userID, &jsonStr); err != nil {
			rows.Close()
			return err
		}

		var emojis []Emoji
		if err := json.Unmarshal([]byte(jsonStr), &emojis); err != nil {
			rows.Close()
			return err
		}

		var html string
		for _, e := range emojis {
			html += fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, e.ID, e.Char)
		}
		cmDownTasks = append(cmDownTasks, cmDownTask{chatID, userID, html})
	}
	rows.Close()

	for _, t := range cmDownTasks {
		if _, err := tx.ExecContext(ctx, `UPDATE chat_members SET emoji = $1 WHERE chat_id = $2 AND user_id = $3`, t.emojiHTML, t.chatID, t.userID); err != nil {
			return err
		}
	}

	type uDownTask struct {
		id        int64
		emojiHTML string
	}
	var uDownTasks []uDownTask

	rows2, err := tx.QueryContext(ctx, `SELECT id, emoji_json FROM users WHERE emoji_json IS NOT NULL`)
	if err != nil {
		return err
	}

	for rows2.Next() {
		var id int64
		var jsonStr string
		if err := rows2.Scan(&id, &jsonStr); err != nil {
			rows2.Close()
			return err
		}

		var emojis []Emoji
		if err := json.Unmarshal([]byte(jsonStr), &emojis); err != nil {
			rows2.Close()
			return err
		}

		var html string
		for _, e := range emojis {
			html += fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, e.ID, e.Char)
		}
		uDownTasks = append(uDownTasks, uDownTask{id, html})
	}
	rows2.Close()

	for _, t := range uDownTasks {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET emoji = $1 WHERE id = $2`, t.emojiHTML, t.id); err != nil {
			return err
		}
	}

	return nil
}

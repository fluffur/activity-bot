package migrations

import (
	"context"
	"database/sql"
	"math/rand/v2"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(UP00059, DOWN00059)
}

var randomEmojis = [...]string{
	"⭐", "🌟", "✨", "💫", "☀️", "🌙", "☄️",

	"🍀", "🌿", "🌱", "🌾", "🌸", "🌺", "🌻", "🌼", "🌷", "🌹", "🪻", "🌵", "🍁", "🍂", "🍃",

	"🌈", "❄️", "⚡", "🔥", "💧", "🌊",

	"💎", "🔹", "🔷", "🪙", "🧿",

	"🪐", "🌌", "🌠", "🛰️", "🚀",

	"🦊", "🐺", "🐱", "🐼", "🐨", "🦁", "🐯", "🐮",
	"🐸", "🐧", "🦉", "🦅", "🦜", "🦢", "🦋", "🐝",
	"🐞", "🦝", "🦔", "🐿️", "🦦", "🐬", "🐳", "🐢",
	"🦕", "🦖", "🦄",

	"🍎", "🍏", "🍐", "🍊", "🍋", "🍉", "🍇",
	"🍓", "🫐", "🍒", "🥝", "🥥", "🥑",

	"☕", "🫖", "🍪", "🍩", "🧁", "🍰", "🍫", "🍯",

	"🎨", "🎭", "🎵", "🎶", "🎸", "🎹",
	"📚", "📖", "✏️", "🖋️", "🧩", "♟️",

	"🎯", "🎲", "🎮", "🕹️", "🏆", "🥇",

	"⚙️", "🛠️", "💡", "🔋", "💻", "⌨️",

	"❤️", "🩵", "💚", "💜", "🖤", "🤍",
	"🔰", "⭕", "🔶", "🔷", "🌀", "♻️",

	"🎷", "🎺", "🥁", "🎻",

	"🪄", "🧭", "🕯️", "🪶", "🔮", "📌", "📎", "🧸",
}

func randomEmoji() string {
	return randomEmojis[rand.IntN(len(randomEmojis))]
}

func UP00059(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM users
		WHERE emoji IS NULL OR emoji = ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []int64

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return err
		}

		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range userIDs {
		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET emoji = $1
			WHERE id = $2
		`, randomEmoji(), id); err != nil {
			return err
		}
	}

	return nil
}

func DOWN00059(context.Context, *sql.Tx) error {
	return nil
}

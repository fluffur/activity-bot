package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"activity-bot/internal/rp"
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type RPRepository struct {
	queries *db.Queries
}

func NewRPRepository(queries *db.Queries) rp.Repository {
	return &RPRepository{queries: queries}
}

func normalizeRPTrigger(trigger string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(trigger)), " "))
}

func (r *RPRepository) Upsert(ctx context.Context, cmd model.RPCommand) error {
	emojiJSON, err := json.Marshal(cmd.Emoji)
	if err != nil {
		return err
	}

	return r.queries.UpsertRPCommand(ctx, db.UpsertRPCommandParams{
		ChatID:            cmd.ChatID,
		Trigger:           cmd.Trigger,
		TriggerNormalized: normalizeRPTrigger(cmd.Trigger),
		Template: pgtype.Text{
			String: cmd.Template,
			Valid:  strings.TrimSpace(cmd.Template) != "",
		},
		EmojiJson: emojiJSON,
		CreatedBy: cmd.CreatedBy,
	})
}

func (r *RPRepository) Delete(ctx context.Context, chatID int64, trigger string) error {
	return r.queries.DeleteRPCommand(ctx, db.DeleteRPCommandParams{
		ChatID:            chatID,
		TriggerNormalized: normalizeRPTrigger(trigger),
	})
}

func (r *RPRepository) GetByTrigger(ctx context.Context, chatID int64, trigger string) (model.RPCommand, error) {
	row, err := r.queries.GetRPCommandByTrigger(ctx, db.GetRPCommandByTriggerParams{
		ChatID:            chatID,
		TriggerNormalized: normalizeRPTrigger(trigger),
	})
	if err != nil {
		return model.RPCommand{}, err
	}
	return mapRPCommand(row), nil
}

func (r *RPRepository) ListByChat(ctx context.Context, chatID int64) ([]model.RPCommand, error) {
	rows, err := r.queries.ListRPCommandsByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return mapMany(rows, mapRPCommand), nil
}

func mapRPCommand(row db.ChatRpCommand) model.RPCommand {
	var emojis model.Emojis
	if len(row.EmojiJson) > 0 {
		_ = json.Unmarshal(row.EmojiJson, &emojis)
	}

	return model.RPCommand{
		ChatID:    row.ChatID,
		Trigger:   row.Trigger,
		Template:  row.Template.String,
		Emoji:     emojis,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

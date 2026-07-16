package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/rp"
	"context"
	"database/sql"
	"strings"
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

func (r *RPRepository) Upsert(ctx context.Context, cmd rp.Definition) error {
	return r.queries.UpsertRPCommand(ctx, db.UpsertRPCommandParams{
		ChatID:    cmd.ChatID,
		Trigger:   normalizeRPTrigger(cmd.Trigger),
		Action:    cmd.Action,
		Emojis:    cmd.Emojis,
		CreatedBy: cmd.CreatedBy,
	})
}

func (r *RPRepository) Delete(ctx context.Context, chatID int64, trigger string) error {
	return r.queries.DeleteRPCommand(ctx, db.DeleteRPCommandParams{
		ChatID:  chatID,
		Trigger: normalizeRPTrigger(trigger),
	})
}

func (r *RPRepository) Get(ctx context.Context, chatID int64, trigger string) (rp.Definition, error) {
	row, err := r.queries.GetRPCommandByTrigger(ctx, db.GetRPCommandByTriggerParams{
		ChatID:  chatID,
		Trigger: normalizeRPTrigger(trigger),
	})
	if err != nil {
		return rp.Definition{}, err
	}
	return mapRPCommand(row), nil
}

func (r *RPRepository) List(ctx context.Context, chatID int64) ([]rp.Definition, error) {
	rows, err := r.queries.ListRPCommandsByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return mapList(rows, mapRPCommand), nil
}

func (r *RPRepository) Match(ctx context.Context, chatID int64, text string) (rp.Definition, int, error) {
	cmds, err := r.List(ctx, chatID)
	if err != nil {
		return rp.Definition{}, 0, err
	}

	text = normalizeRPTrigger(text)

	var (
		best    rp.Definition
		bestLen int
		found   bool
	)

	for _, cmd := range cmds {
		trigger := normalizeRPTrigger(cmd.Trigger)

		if !strings.HasPrefix(text, trigger) {
			continue
		}

		if len(text) > len(trigger) {
			switch text[len(trigger)] {
			case ' ', '\n', '\t':
			default:
				continue
			}
		}

		if len(trigger) > bestLen {
			best = cmd
			bestLen = len(trigger)
			found = true
		}
	}

	if !found {
		return rp.Definition{}, 0, sql.ErrNoRows
	}

	return best, bestLen, nil
}

func mapRPCommand(row db.ChatRpCommand) rp.Definition {
	return rp.Definition{
		ChatID:    row.ChatID,
		Trigger:   row.Trigger,
		Action:    row.Action,
		Emojis:    row.Emojis,
		CreatedBy: row.CreatedBy,
	}
}

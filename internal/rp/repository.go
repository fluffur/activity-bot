package rp

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Upsert(ctx context.Context, cmd model.RPCommand) error
	Delete(ctx context.Context, chatID int64, trigger string) error
	GetByTrigger(ctx context.Context, chatID int64, trigger string) (model.RPCommand, error)
	ListByChat(ctx context.Context, chatID int64) ([]model.RPCommand, error)
}

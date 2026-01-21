package message

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Save(ctx context.Context, m model.Message) error
}

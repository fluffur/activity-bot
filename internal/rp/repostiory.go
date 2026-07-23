package rp

import (
	"context"
)

type Repository interface {
	Get(ctx context.Context, chatID int64, trigger string) (Definition, error)
	GetByID(ctx context.Context, id int64) (Definition, error)
	List(ctx context.Context, chatID int64) ([]Definition, error)
	Upsert(ctx context.Context, cmd Definition) error
	Delete(ctx context.Context, chatID int64, trigger string) error
	Match(ctx context.Context, chatID int64, text string) (Definition, int, error)
}

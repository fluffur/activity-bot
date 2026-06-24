package norm

import "context"

type Repository interface {
	Get(ctx context.Context, chatID int64, name string) (Norm, error)
	Set(ctx context.Context, chatID int64, name string, value int32) error
	List(ctx context.Context, chatID int64) ([]Norm, error)
}

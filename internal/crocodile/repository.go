package crocodile

import "context"

type Repository interface {
	Create(ctx context.Context, game *Game) error
	Get(ctx context.Context, chatID int64) (*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, chatID int64) error

	Exists(ctx context.Context, chatID int64) (bool, error)
}

type WordRepository interface {
	GetRandom(ctx context.Context) (Word, error)
	MarkUsed(ctx context.Context, id int64) error
}

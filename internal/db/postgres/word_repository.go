package postgres

import (
	"activity-bot/internal/crocodile"
	db "activity-bot/internal/db/postgres/sqlc"
	"context"
)

type WordRepository struct {
	queries *db.Queries
}

func NewWordRepository(
	queries *db.Queries,
) *WordRepository {
	return &WordRepository{
		queries: queries,
	}
}

func (r *WordRepository) GetRandom(
	ctx context.Context,
) (crocodile.Word, error) {
	w, err := r.queries.GetRandomCrocodileWord(ctx)
	if err != nil {
		return crocodile.Word{}, err
	}

	return crocodile.Word{
		ID:         w.ID,
		Word:       w.Word,
		Category:   w.Category,
		Difficulty: w.Difficulty,
		UsedCount:  w.UsedCount,
	}, nil
}

func (r *WordRepository) MarkUsed(
	ctx context.Context,
	id int64,
) error {
	return r.queries.MarkCrocodileWordUsed(
		ctx,
		id,
	)
}

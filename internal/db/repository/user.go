package repository

import (
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/user"
	"context"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) user.Repository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, user user.User) error {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:        user.ID,
		Username:  text(user.Username),
		FirstName: text(user.FirstName),
		LastName:  text(user.LastName),
		CreatedAt: timestamptz(user.CreatedAt),
		Gender:    string(user.Gender),
		IsBot:     user.IsBot,
	})
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (user.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return user.User{}, err
	}
	return mapUser(u), nil
}

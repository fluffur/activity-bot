package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/user"
	"context"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) user.Repository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) error {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:        u.ID,
		Username:  text(u.Username),
		FirstName: text(u.FirstName),
		LastName:  text(u.LastName),
		CreatedAt: timestamptz(u.CreatedAt),
		Gender:    string(u.Gender),
		IsBot:     u.IsBot,
	})
}

func (r *UserRepository) Get(ctx context.Context, id int64) (user.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return user.User{}, err
	}

	return mapUser(u), nil
}

func (r *UserRepository) Update(ctx context.Context, u user.User) error {
	return r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        u.ID,
		Username:  text(u.Username),
		FirstName: text(u.FirstName),
		LastName:  text(u.LastName),
	})
}

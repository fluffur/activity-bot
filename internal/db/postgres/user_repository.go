package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries db.Querier
}

func NewUserRepository(queries db.Querier) *UserRepository {
	return &UserRepository{queries}
}

func (r *UserRepository) EnsureExists(ctx context.Context, id int64, username, firstName, lastName string) error {
	return r.queries.EnsureUserExists(ctx, db.EnsureUserExistsParams{
		ID: id,
		Username: pgtype.Text{
			String: username,
			Valid:  username != "",
		},
		FirstName: pgtype.Text{
			String: firstName,
			Valid:  firstName != "",
		},
		LastName: pgtype.Text{
			String: lastName,
			Valid:  lastName != "",
		},
	})
}

func (r *UserRepository) Get(ctx context.Context, id int64) (model.User, error) {
	u, err := r.queries.GetUser(ctx, id)
	if err != nil {
		return model.User{}, err
	}

	return mapUser(u), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	u, err := r.queries.GetUserByUsername(ctx, pgtype.Text{
		String: username,
		Valid:  username != "",
	})
	if err != nil {
		return model.User{}, err
	}

	return mapUser(u), nil
}

func mapUser(u db.User) model.User {
	nu := model.User{
		ID: u.ID,
	}

	if u.FirstName.Valid {
		nu.FirstName = u.FirstName.String
	}
	if u.LastName.Valid {
		nu.LastName = u.LastName.String
	}
	if u.Username.Valid {
		nu.Username = &u.Username.String
	}
	return nu
}

package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) user.Repository {
	return &UserRepository{queries}
}

func (r *UserRepository) Ensure(ctx context.Context, id int64, username, firstName, lastName string) (model.User, error) {
	u, err := r.queries.EnsureUserExists(ctx, db.EnsureUserExistsParams{
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
	if err != nil {
		return model.User{}, err
	}

	return mapUser(u), nil
}

func (r *UserRepository) Get(ctx context.Context, id int64) (model.User, error) {
	u, err := r.queries.GetUser(ctx, id)
	if err != nil {
		return model.User{}, err
	}

	return mapUser(u), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	u, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return model.User{}, err
	}

	return mapUser(u), nil
}

func (r *UserRepository) UpsertUsers(ctx context.Context, users []model.User) error {
	ids := make([]int64, len(users))
	usernames := make([]string, len(users))
	firstNames := make([]string, len(users))
	lastNames := make([]string, len(users))

	for i, u := range users {
		ids[i] = u.ID
		if u.Username != nil {
			usernames[i] = *u.Username
		}
		firstNames[i] = u.FirstName
		lastNames[i] = u.LastName
	}

	return r.queries.UpsertUsers(ctx, db.UpsertUsersParams{
		Ids:        ids,
		Usernames:  usernames,
		FirstNames: firstNames,
		LastNames:  lastNames,
	})
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

func (r *UserRepository) GetByCustomTitle(ctx context.Context, chatID int64, title string) ([]model.ChatMember, error) {
	u, err := r.queries.GetUsersByCustomTitle(ctx, db.GetUsersByCustomTitleParams{
		CustomTitle: pgtype.Text{
			String: title,
			Valid:  true,
		},
		ChatID: chatID,
	})

	results := make([]model.ChatMember, len(u))
	for i, u := range u {
		results[i] = mapChatMemberRow(db.GetChatMemberRow(u))
	}
	return results, err

}

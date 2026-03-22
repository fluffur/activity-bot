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
		if u.Username != "" {
			usernames[i] = u.Username
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

func (r *UserRepository) GetByTag(ctx context.Context, chatID int64, tag string) ([]model.ChatMember, error) {
	u, err := r.queries.GetUsersByCustomTitle(ctx, db.GetUsersByCustomTitleParams{
		Tag:    tag,
		ChatID: chatID,
	})

	return mapMembers(u, func(row db.GetUsersByCustomTitleRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), err
}

func (r *UserRepository) SetGender(ctx context.Context, userID int64, gender string) error {
	return r.queries.SetUserGender(ctx, db.SetUserGenderParams{
		ID:     userID,
		Gender: gender,
	})
}

func (r *UserRepository) SetEmoji(ctx context.Context, userID int64, emoji string) error {
	return r.queries.SetUserEmoji(ctx, db.SetUserEmojiParams{
		ID: userID,
		Emoji: pgtype.Text{
			String: emoji,
			Valid:  emoji != "",
		},
	})
}

func (r *UserRepository) SetCustomEmojiID(ctx context.Context, userID int64, emojiID string) error {
	return r.queries.SetUserCustomEmojiID(ctx, db.SetUserCustomEmojiIDParams{
		ID: userID,
		CustomEmojiID: pgtype.Text{
			String: emojiID,
			Valid:  emojiID != "",
		},
	})
}

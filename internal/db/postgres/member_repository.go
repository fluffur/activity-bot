package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MemberRepository struct {
	queries *db.Queries
}

func NewMemberRepository(queries *db.Queries) member.Repository {
	return &MemberRepository{queries}
}

func (r *MemberRepository) GetCustomTitle(ctx context.Context, chatID int64, userID int64) (string, error) {
	title, err := r.queries.GetMemberCustomTitle(ctx, db.GetMemberCustomTitleParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return "", member.ErrMemberNotFound
	}
	if err != nil {
		return "", err
	}
	if !title.Valid || title.String == "" {
		return "", member.ErrInvalidCustomTitle
	}

	return title.String, nil
}

func (r *MemberRepository) UpdateCustomTitle(ctx context.Context, chatID int64, userID int64, title *string) error {
	var customTitle pgtype.Text
	if title != nil {
		customTitle = pgtype.Text{
			String: *title,
			Valid:  true,
		}
	}
	return r.queries.UpdateChatMemberTitle(ctx, db.UpdateChatMemberTitleParams{
		ChatID:      chatID,
		UserID:      userID,
		CustomTitle: customTitle,
	})

}

func (r *MemberRepository) UpdateRole(ctx context.Context, chatID int64, userID int64, role string) error {
	return r.queries.UpdateChatMemberRole(ctx, db.UpdateChatMemberRoleParams{
		ChatID: chatID,
		UserID: userID,
		Role:   role,
	})
}

func (r *MemberRepository) FindByChatID(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	members, err := r.queries.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]model.ChatMember, len(members))
	for i, m := range members {
		var exemptUntil *time.Time
		var username *string
		if m.Username.Valid {
			username = &m.Username.String
		}
		if m.ExemptUntil.Valid {
			exemptUntil = &m.ExemptUntil.Time
		}
		result[i] = model.ChatMember{
			User: model.User{
				ID:        m.UserID,
				FirstName: m.FirstName.String,
				LastName:  m.LastName.String,
				Username:  username,
			},
			ChatID:      chatID,
			ExemptUntil: exemptUntil,
			CustomTitle: m.CustomTitle.String,
			Role:        m.Role,
		}
	}

	return result, nil
}

func (r *MemberRepository) GetWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	members, err := r.queries.GetChatMembersWithTitles(ctx, chatID)
	if err != nil {
		return nil, err
	}
	res := make([]model.ChatMember, len(members))
	for i, m := range members {
		var username *string
		if m.Username.Valid {
			username = &m.Username.String
		}

		res[i] = model.ChatMember{
			ChatID: chatID,
			User: model.User{
				Username:  username,
				FirstName: m.FirstName.String,
				LastName:  m.LastName.String,
				ID:        m.UserID,
			},
			CustomTitle: m.CustomTitle.String,
			Role:        m.Role,
		}
	}
	return res, nil
}

func (r *MemberRepository) UpsertChatMembers(ctx context.Context, chatID int64, users []model.ChatMemberUpdate) error {
	userIDs := make([]int64, len(users))
	customTitles := make([]string, len(users))
	roles := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.User.ID
		customTitles[i] = u.CustomTitle
		roles[i] = u.Role
	}

	return r.queries.UpsertChatMembersWithRole(ctx, db.UpsertChatMembersWithRoleParams{
		ChatID:       chatID,
		UserIds:      userIDs,
		CustomTitles: customTitles,
		Roles:        roles,
	})
}

func (r *MemberRepository) Get(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapChatMemberRow(m), nil
}

func (r *MemberRepository) Remove(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.DeleteChatMember(ctx, db.DeleteChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *MemberRepository) EnsureExists(ctx context.Context, chatID int64, userID int64, role string) (model.ChatMember, error) {
	m, err := r.queries.EnsureChatMemberExists(ctx, db.EnsureChatMemberExistsParams{
		ChatID: chatID,
		UserID: userID,
		Role:   role,
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapChatMember(m), nil
}

func (r *MemberRepository) EnsureFull(ctx context.Context, chatID int64, userID int64, role string, firstName, lastName string, username string, weeklyNorm int32) (model.ChatMember, error) {
	m, err := r.queries.EnsureMemberFull(ctx, db.EnsureMemberFullParams{
		Role:       role,
		ChatID:     chatID,
		WeeklyNorm: weeklyNorm,
		UserID:     userID,
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
		return model.ChatMember{}, err
	}

	return mapChatMember(m), nil
}

func (r *MemberRepository) MarkLeftNotInList(ctx context.Context, chatID int64, userIDs []int64) error {
	return r.queries.MarkChatMembersLeftNotInList(ctx, db.MarkChatMembersLeftNotInListParams{
		ChatID:  chatID,
		UserIds: userIDs,
	})
}

func mapChatMember(m db.ChatMember) model.ChatMember {
	var exemptUntil *time.Time
	if m.ExemptUntil.Valid {
		t := m.ExemptUntil.Time
		exemptUntil = &t
	}
	return model.ChatMember{
		ChatID:      m.ChatID,
		User:        model.User{ID: m.UserID},
		ExemptUntil: exemptUntil,
		CustomTitle: m.CustomTitle.String,
		Role:        m.Role,
	}
}

func mapChatMemberRow(m db.GetChatMemberRow) model.ChatMember {
	var exemptUntil *time.Time
	if m.ExemptUntil.Valid {
		t := m.ExemptUntil.Time
		exemptUntil = &t
	}
	var username *string
	if m.Username.Valid {
		username = &m.Username.String
	}
	return model.ChatMember{
		ChatID: m.ChatID,
		User: model.User{
			ID:        m.UserID,
			FirstName: m.FirstName.String,
			LastName:  m.LastName.String,
			Username:  username,
		},
		ExemptUntil: exemptUntil,
		CustomTitle: m.CustomTitle.String,
		Role:        m.Role,
	}
}

func (r *MemberRepository) SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	return r.queries.MoveChatMembersToOldExcept(ctx, db.MoveChatMembersToOldExceptParams{
		ChatID:  chatID,
		UserIds: userIDs,
	})
}

func (r *MemberRepository) SetNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	log.Println(userIDs)
	return r.queries.MoveChatMembersToNew(ctx, db.MoveChatMembersToNewParams{
		ChatID:  chatID,
		UserIds: userIDs,
	})
}

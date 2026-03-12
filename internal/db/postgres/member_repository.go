package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NormMode string

const (
	NormModeWarn NormMode = "warn"
	NormModeBan  NormMode = "ban"
	NormModeAny  NormMode = "any"
)

type MemberRepository struct {
	queries *db.Queries
}

func (r *MemberRepository) GetNoNormWarnMembers(ctx context.Context, id int64, from, to *time.Time) ([]model.ChatMember, error) {
	members, err := r.queries.GetNoNormMembers(ctx, noNormMembersParams(id, from, to, NormModeWarn))
	if err != nil {
		return nil, err
	}
	return mapMembers(members), nil
}

func (r *MemberRepository) GetNoNormBanMembers(ctx context.Context, id int64, from, to *time.Time) ([]model.ChatMember, error) {
	members, err := r.queries.GetNoNormMembers(ctx, noNormMembersParams(id, from, to, NormModeBan))
	if err != nil {
		return nil, err
	}
	return mapMembers(members), nil

}
func (r *MemberRepository) GetNoNormMembers(ctx context.Context, id int64, from, to *time.Time) ([]model.ChatMember, error) {
	members, err := r.queries.GetNoNormMembers(ctx, noNormMembersParams(id, from, to, NormModeAny))
	if err != nil {
		return nil, err
	}
	return mapMembers(members), nil
}

func noNormMembersParams(id int64, from, to *time.Time, mode NormMode) db.GetNoNormMembersParams {
	var fromTime time.Time
	if from != nil {
		fromTime = *from
	}

	var toTime time.Time
	if to != nil {
		toTime = *to
	}
	return db.GetNoNormMembersParams{
		FromDate: pgtype.Timestamptz{
			Time:  fromTime,
			Valid: from != nil,
		},
		ToDate: pgtype.Timestamptz{
			Time:  toTime,
			Valid: to != nil,
		},
		ChatID: id,
		Mode:   mode,
	}
}

func mapMembers(members []db.GetNoNormMembersRow) []model.ChatMember {
	result := make([]model.ChatMember, len(members))
	for i, m := range members {
		var restUntil *time.Time
		if m.RestUntil.Valid {
			restUntil = &m.RestUntil.Time
		}
		var username *string
		if m.Username.Valid {
			username = &m.Username.String
		}
		result[i] = model.ChatMember{
			User: model.User{
				ID:            m.ID,
				FirstName:     m.FirstName.String,
				LastName:      m.LastName.String,
				Username:      username,
				Gender:        m.Gender,
				Emoji:         m.Emoji.String,
				CustomEmojiID: m.CustomEmojiID.String,
			},
			ChatID:      m.ChatID,
			RestUntil:   restUntil,
			RestReason:  m.RestReason.String,
			CustomTitle: m.CustomTitle.String,
			Status:      m.Status,
		}
	}
	return result
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

func (r *MemberRepository) UpdateStatus(ctx context.Context, chatID int64, userID int64, status string) error {
	return r.queries.UpdateMemberStatus(ctx, db.UpdateMemberStatusParams{
		ChatID: chatID,
		UserID: userID,
		Status: status,
	})
}

func (r *MemberRepository) FindByChatID(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	members, err := r.queries.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]model.ChatMember, len(members))
	for i, m := range members {
		var restUntil *time.Time
		var username *string
		if m.Username.Valid {
			username = &m.Username.String
		}
		if m.RestUntil.Valid {
			restUntil = &m.RestUntil.Time
		}
		result[i] = model.ChatMember{
			User: model.User{
				ID:            m.UserID,
				FirstName:     m.FirstName.String,
				LastName:      m.LastName.String,
				Username:      username,
				Gender:        m.Gender,
				Emoji:         m.Emoji.String,
				CustomEmojiID: m.CustomEmojiID.String,
			},
			ChatID:      chatID,
			RestUntil:   restUntil,
			CustomTitle: m.CustomTitle.String,
			Status:      m.Status,
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
				ID:            m.UserID,
				FirstName:     m.FirstName.String,
				LastName:      m.LastName.String,
				Username:      username,
				Gender:        m.Gender,
				Emoji:         m.Emoji.String,
				CustomEmojiID: m.CustomEmojiID.String,
			},
			CustomTitle: m.CustomTitle.String,
			Status:      m.Status,
		}
	}
	return res, nil
}

func (r *MemberRepository) GetAnyWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	members, err := r.queries.GetAnyChatMembersWithTitles(ctx, chatID)
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
				ID:            m.UserID,
				FirstName:     m.FirstName.String,
				LastName:      m.LastName.String,
				Username:      username,
				Gender:        m.Gender,
				Emoji:         m.Emoji.String,
				CustomEmojiID: m.CustomEmojiID.String,
			},
			CustomTitle: m.CustomTitle.String,
			Status:      m.Status,
		}
	}
	return res, nil
}

func (r *MemberRepository) UpsertChatMembers(ctx context.Context, chatID int64, users []model.ChatMemberUpdate) error {
	userIDs := make([]int64, len(users))
	customTitles := make([]string, len(users))
	statuses := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.User.ID
		customTitles[i] = u.CustomTitle
		statuses[i] = u.Status
	}

	return r.queries.UpsertChatMembers(ctx, db.UpsertChatMembersParams{
		ChatID:       chatID,
		UserIds:      userIDs,
		CustomTitles: customTitles,
		Statuses:     statuses,
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

func (r *MemberRepository) EnsureExists(ctx context.Context, chatID int64, userID int64, status string) (model.ChatMember, error) {
	m, err := r.queries.EnsureChatMemberExists(ctx, db.EnsureChatMemberExistsParams{
		ChatID: chatID,
		UserID: userID,
		Status: status,
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapChatMember(m), nil
}

func (r *MemberRepository) EnsureFull(ctx context.Context, chatID, userID int64, role, firstName, lastName, username string, normWarn int32) (model.ChatMember, error) {
	m, err := r.queries.EnsureMemberFull(ctx, db.EnsureMemberFullParams{
		CustomTitle: pgtype.Text{
			String: role,
			Valid:  true,
		},
		ChatID:   chatID,
		NormWarn: normWarn,
		UserID:   userID,
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
	var restUntil *time.Time
	if m.RestUntil.Valid {
		t := m.RestUntil.Time
		restUntil = &t
	}
	return model.ChatMember{
		ChatID: m.ChatID,
		User: model.User{
			ID: m.UserID,
		},
		RestUntil:   restUntil,
		CustomTitle: m.CustomTitle.String,
		Status:      m.Status,
	}
}

func mapChatMemberRow(m db.GetChatMemberRow) model.ChatMember {
	var restUntil *time.Time
	if m.RestUntil.Valid {
		t := m.RestUntil.Time
		restUntil = &t
	}
	var username *string
	if m.Username.Valid {
		username = &m.Username.String
	}
	return model.ChatMember{
		ChatID: m.ChatID,
		User: model.User{
			ID:            m.UserID,
			FirstName:     m.FirstName.String,
			LastName:      m.LastName.String,
			Username:      username,
			Gender:        m.Gender,
			Emoji:         m.Emoji.String,
			CustomEmojiID: m.CustomEmojiID.String,
		},
		RestUntil:   restUntil,
		RestReason:  m.RestReason.String,
		CustomTitle: m.CustomTitle.String,
		Status:      m.Status,
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
	return r.queries.MoveChatMembersToNew(ctx, db.MoveChatMembersToNewParams{
		ChatID:  chatID,
		UserIds: userIDs,
	})
}

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
	return mapMembers(members, func(row db.GetNoNormMembersRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), nil
}

func (r *MemberRepository) GetNoNormBanMembers(ctx context.Context, id int64, from, to *time.Time) ([]model.ChatMember, error) {
	members, err := r.queries.GetNoNormMembers(ctx, noNormMembersParams(id, from, to, NormModeBan))
	if err != nil {
		return nil, err
	}
	return mapMembers(members, func(row db.GetNoNormMembersRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), nil

}
func (r *MemberRepository) GetNoNormMembers(ctx context.Context, id int64, from, to *time.Time) ([]model.ChatMember, error) {
	members, err := r.queries.GetNoNormMembers(ctx, noNormMembersParams(id, from, to, NormModeAny))
	if err != nil {
		return nil, err
	}
	return mapMembers(members, func(row db.GetNoNormMembersRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), nil
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

func NewMemberRepository(queries *db.Queries) member.Repository {
	return &MemberRepository{queries}
}

func (r *MemberRepository) GetTag(ctx context.Context, chatID int64, userID int64) (string, error) {
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

func (r *MemberRepository) UpdateCustomTitle(ctx context.Context, chatID int64, userID int64, title string) error {
	return r.queries.UpdateChatMemberTitle(ctx, db.UpdateChatMemberTitleParams{
		ChatID: chatID,
		UserID: userID,
		Tag: pgtype.Text{
			String: title,
			Valid:  title != "",
		},
	})

}

func (r *MemberRepository) UpdateStatus(ctx context.Context, chatID int64, userID int64, status int16) error {
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
		result[i] = mapChatMemberFull(m.ChatMember, m.User)
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
		res[i] = mapChatMemberFull(m.ChatMember, m.User)
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
		res[i] = mapChatMemberFull(m.ChatMember, m.User)
	}
	return res, nil
}

func (r *MemberRepository) UpsertChatMembers(ctx context.Context, chatID int64, users []model.ChatMemberUpdate) error {
	userIDs := make([]int64, len(users))
	tags := make([]string, len(users))
	statuses := make([]int16, len(users))
	for i, u := range users {
		userIDs[i] = u.User.ID
		tags[i] = u.Tag
		statuses[i] = u.Status
	}

	return r.queries.UpsertChatMembers(ctx, db.UpsertChatMembersParams{
		ChatID:   chatID,
		UserIds:  userIDs,
		Tags:     tags,
		Statuses: statuses,
	})
}

func (r *MemberRepository) GetChatMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapChatMemberFull(m.ChatMember, m.User), nil
}

func (r *MemberRepository) Remove(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.DeleteChatMember(ctx, db.DeleteChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *MemberRepository) EnsureFull(ctx context.Context, chatID, userID int64, role, firstName, lastName, username string) (model.ChatMember, error) {
	m, err := r.queries.EnsureMemberFull(ctx, db.EnsureMemberFullParams{
		Tag: pgtype.Text{
			String: role,
			Valid:  role != "",
		},
		ChatID: chatID,
		UserID: userID,
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

	return mapChatMemberFull(m.ChatMember, m.User), nil
}

func (r *MemberRepository) MarkLeftNotInList(ctx context.Context, chatID int64, userIDs []int64) error {
	return r.queries.MarkChatMembersLeftNotInList(ctx, db.MarkChatMembersLeftNotInListParams{
		ChatID:  chatID,
		UserIds: userIDs,
	})
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

func (r *MemberRepository) FindByTag(ctx context.Context, chatID int64, tag string) ([]model.ChatMember, error) {
	m, err := r.queries.FindChatMembersByTag(ctx, db.FindChatMembersByTagParams{
		ChatID: chatID,
		Tag:    tag,
	})
	if err != nil {
		return nil, err
	}

	return mapMembers(m, func(row db.FindChatMembersByTagRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), err
}

func (r *MemberRepository) GetByUsername(ctx context.Context, chatID int64, username string) (model.ChatMember, error) {
	m, err := r.queries.GetChatMemberByUsername(ctx, db.GetChatMemberByUsernameParams{
		ChatID:   chatID,
		Username: pgtype.Text{String: username, Valid: true},
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapChatMemberFull(m.ChatMember, m.User), nil
}

func (r *MemberRepository) SetEmoji(ctx context.Context, chatID, userID int64, emojis model.Emojis) error {
	return r.queries.SetChatMemberEmojiJSON(ctx, db.SetChatMemberEmojiJSONParams{
		ChatID:    chatID,
		UserID:    userID,
		EmojiJson: emojis,
	})
}

func (r *MemberRepository) ResetCreators(ctx context.Context, chatID int64) error {
	return r.queries.ResetChatMemberCreators(ctx, chatID)
}

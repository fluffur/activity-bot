package postgres

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/permission"
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatMemberRepository struct {
	queries *db.Queries
}

func NewChatMemberRepository(queries *db.Queries) chatmember.Repository {
	return &ChatMemberRepository{queries: queries}
}

func (r *ChatMemberRepository) Create(ctx context.Context, m chatmember.ChatMember) error {
	return r.queries.CreateChatMember(ctx, db.CreateChatMemberParams{
		ChatID:     m.Chat.ID,
		UserID:     m.User.ID,
		Tag:        text(m.Tag),
		Status:     int16(m.Status),
		RestUntil:  timestamptz(m.RestUntil),
		LeftAt:     timestamptz(m.LeftAt),
		RestReason: text(m.RestReason),
	})
}

func (r *ChatMemberRepository) Get(ctx context.Context, chatID, userID int64) (chatmember.ChatMember, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return chatmember.ChatMember{}, err
	}

	return mapChatMemberFull(m.ChatMember, m.Chat, m.User), nil
}

func (r *ChatMemberRepository) SetTag(ctx context.Context, chatID, userID int64, tag string) error {
	return r.queries.SetChatMemberTag(ctx, db.SetChatMemberTagParams{
		Tag:    text(tag),
		UserID: userID,
		ChatID: chatID,
	})
}

func (r *ChatMemberRepository) MarkLeft(ctx context.Context, chatID, userID int64, leftAt time.Time) error {
	return r.queries.MarkChatMemberLeft(ctx, db.MarkChatMemberLeftParams{
		LeftAt: timestamptz(leftAt),
		UserID: userID,
		ChatID: chatID,
	})
}

func (r *ChatMemberRepository) MarkAllLeftExcept(ctx context.Context, chatID int64, userIDs []int64, leftAt time.Time) error {
	return r.queries.MarkAllChatMembersLeftExcept(ctx, db.MarkAllChatMembersLeftExceptParams{
		ChatID:  chatID,
		UserIds: userIDs,
		LeftAt:  timestamptz(leftAt),
	})
}

func (r *ChatMemberRepository) UpsertChatMembers(ctx context.Context, chatID int64, chatMembers []chatmember.ChatMember) error {
	if len(chatMembers) == 0 {
		return nil
	}

	sort.Slice(chatMembers, func(i, j int) bool {
		return chatMembers[i].User.ID < chatMembers[j].User.ID
	})

	n := len(chatMembers)
	userIds := make([]int64, n)
	tags := make([]string, n)
	statuses := make([]int16, n)
	usernames := make([]string, n)
	firstNames := make([]string, n)
	lastNames := make([]string, n)
	isBots := make([]bool, n)

	for i, member := range chatMembers {
		userIds[i] = member.User.ID
		tags[i] = member.Tag
		statuses[i] = int16(member.Status)
		usernames[i] = member.User.Username
		firstNames[i] = member.User.FirstName
		lastNames[i] = member.User.LastName
		isBots[i] = member.User.IsBot
	}

	return r.queries.UpsertChatMembersAndUsers(ctx, db.UpsertChatMembersAndUsersParams{
		ChatID:     chatID,
		UserIds:    userIds,
		Tags:       tags,
		Statuses:   statuses,
		Usernames:  usernames,
		FirstNames: firstNames,
		LastNames:  lastNames,
		IsBots:     isBots,
	})
}

func (r *ChatMemberRepository) Restore(ctx context.Context, chatID, userID int64) error {
	return r.queries.MarkChatMemberLeft(ctx, db.MarkChatMemberLeftParams{
		LeftAt: pgtype.Timestamptz{},
		UserID: userID,
		ChatID: chatID,
	})
}

func (r *ChatMemberRepository) List(ctx context.Context, filter chatmember.Filter) ([]chatmember.ChatMember, error) {
	params := db.ListChatMembersParams{
		ChatID: filter.ChatID,
		IsBot: pgtype.Bool{
			Bool:  filter.IsBot.Bool,
			Valid: filter.IsBot.Valid,
		},
		HasLeft: pgtype.Bool{
			Bool:  filter.Left.Bool,
			Valid: filter.Left.Valid,
		},
		ExcludeFromCall: pgtype.Bool{
			Bool:  filter.Excluded.Bool,
			Valid: filter.Excluded.Valid,
		},
	}

	cms, err := r.queries.ListChatMembers(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapList(cms, func(r db.ListChatMembersRow) chatmember.ChatMember {
		return mapChatMemberFull(r.ChatMember, db.Chat{}, r.User)
	}), nil
}

func (r *ChatMemberRepository) GetByUsername(ctx context.Context, chatID int64, username string) (chatmember.ChatMember, error) {
	cm, err := r.queries.GetChatMemberByUsername(ctx, db.GetChatMemberByUsernameParams{
		ChatID:   chatID,
		Username: text(username),
	})
	if err != nil {
		return chatmember.ChatMember{}, err
	}

	return mapChatMemberFull(cm.ChatMember, db.Chat{}, cm.User), nil
}

func (r *ChatMemberRepository) SetExcludeFromSummon(ctx context.Context, chatID, userID int64, excluded bool) error {
	return r.queries.SetChatMemberExcludeFromSummon(ctx, db.SetChatMemberExcludeFromSummonParams{
		ExcludeFromCall: excluded,
		UserID:          userID,
		ChatID:          chatID,
	})
}

func (r *ChatMemberRepository) ListAdmins(ctx context.Context, chatID int64, minStatus permission.Status) ([]chatmember.ChatMember, error) {
	cms, err := r.queries.ListChatAdmins(ctx, db.ListChatAdminsParams{
		ChatID: chatID,
		Status: int16(minStatus),
	})
	if err != nil {
		return nil, err
	}

	return mapList(cms, func(r db.ListChatAdminsRow) chatmember.ChatMember {
		return mapChatMemberFull(r.ChatMember, db.Chat{}, r.User)
	}), nil
}

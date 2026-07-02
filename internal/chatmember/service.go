package chatmember

import (
	"context"
	"errors"
	"fmt"
	"time"

	"activity-bot/internal/chat"
	"activity-bot/internal/user"

	"github.com/gotd/botapi"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	chatRepo       chat.Repository
	userRepo       user.Repository
	chatMemberRepo Repository
}

func NewService(cr chat.Repository, ur user.Repository, cmr Repository) *Service {
	return &Service{
		chatRepo:       cr,
		userRepo:       ur,
		chatMemberRepo: cmr,
	}
}

type JoinResult struct {
	ChatMember ChatMember
	IsNew      bool
}

func (s *Service) HandleJoin(
	ctx context.Context,
	chatID int64,
	chatTitle string,
	u user.User,
	rank string,
) (JoinResult, error) {
	chatt, err := s.chatRepo.Get(ctx, chatID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return JoinResult{}, fmt.Errorf("service get chat: %w", err)
		}

		chatt = chat.New(chatID, chatTitle)
		if err = s.chatRepo.Create(ctx, chatt); err != nil {
			return JoinResult{}, fmt.Errorf("service create chat: %w", err)
		}
	}

	userr, err := s.userRepo.Get(ctx, u.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return JoinResult{}, fmt.Errorf("service get user: %w", err)
		}

		userr = u
		if err = s.userRepo.Create(ctx, userr); err != nil {
			return JoinResult{}, fmt.Errorf("service create user: %w", err)
		}
	}

	var isNewMember bool

	cm, err := s.chatMemberRepo.Get(ctx, chatID, userr.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return JoinResult{}, fmt.Errorf("service get member: %w", err)
		}

		isNewMember = true
		cm = New(userr, chatt, rank, StatusMember, time.Now())

		if err = s.chatMemberRepo.Create(ctx, cm); err != nil {
			return JoinResult{}, fmt.Errorf("service create member: %w", err)
		}
	}

	if !isNewMember {
		if err = s.chatMemberRepo.Restore(ctx, chatID, userr.ID); err != nil {
			return JoinResult{}, fmt.Errorf("service restore member: %w", err)
		}
	}

	return JoinResult{
		ChatMember: cm,
		IsNew:      isNewMember,
	}, nil
}

func (s *Service) HandleLeft(ctx context.Context, chatID, userID int64) (ChatMember, error) {
	cm, err := s.chatMemberRepo.Get(ctx, chatID, userID)
	if err != nil {
		return ChatMember{}, fmt.Errorf("handle left get member: %w", err)
	}

	if err := s.chatMemberRepo.MarkLeft(ctx, chatID, userID, time.Now()); err != nil {
		return ChatMember{}, fmt.Errorf("service mark left: %w", err)
	}

	return cm, nil
}

func (s *Service) SyncChatMembers(ctx context.Context, chatID int64, cms []ChatMember) error {
	cmIDs := make([]int64, len(cms))
	for i, cm := range cms {
		cmIDs[i] = cm.User.ID
	}

	if err := s.chatMemberRepo.MarkAllLeftExcept(ctx, chatID, cmIDs, time.Now()); err != nil {
		return fmt.Errorf("service sync mark left: %w", err)
	}

	if err := s.chatMemberRepo.UpsertChatMembers(ctx, chatID, cms); err != nil {
		return fmt.Errorf("service sync upsert: %w", err)
	}

	return nil
}

func (s *Service) UpdateTag(ctx context.Context, chatID, userID int64, newRank string) error {
	return s.chatMemberRepo.SetTag(ctx, chatID, userID, newRank)
}

func (s *Service) ListSummonChatMembers(ctx context.Context, chatID int64) ([]ChatMember, error) {
	return s.chatMemberRepo.List(ctx, Filter{
		ChatID: chatID,
		IsBot: OptionalBool{
			Bool:  false,
			Valid: true,
		},
		Left: OptionalBool{
			Bool:  false,
			Valid: true,
		},
		Excluded: OptionalBool{
			Bool:  false,
			Valid: true,
		},
	})
}

func (s *Service) SetExcludeFromSummon(ctx context.Context, chatID, userID int64, excluded bool) error {
	return s.chatMemberRepo.SetExcludeFromSummon(ctx, chatID, userID, excluded)
}

func (s *Service) Get(c *botapi.Context, chatID int64, userID int64) (ChatMember, error) {
	return s.chatMemberRepo.Get(c, chatID, userID)
}

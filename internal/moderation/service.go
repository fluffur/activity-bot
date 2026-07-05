package moderation

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/permission"
	"context"
	"errors"
	"time"
)

var (
	ErrUserStatusInvalid   = errors.New("user status invalid")
	ErrUserCantBeModerated = errors.New("user is protected")
	ErrInvalidRange        = errors.New("invalid range")
)

type Service struct {
	repo           Repository
	chatMemberRepo chatmember.Repository
	ownerID        int64
}

func NewService(
	repo Repository,
	chatMemberRepo chatmember.Repository,
	ownerID int64,
) *Service {
	return &Service{
		repo:           repo,
		chatMemberRepo: chatMemberRepo,
		ownerID:        ownerID,
	}
}

func (s *Service) IsOwner(cm chatmember.ChatMember) bool {
	return cm.ID() == s.ownerID
}

func (s *Service) SetStatus(ctx context.Context, chatID int64, sender, m chatmember.ChatMember, status permission.Status) error {
	if !sender.CanModerate(m) {
		return ErrUserCantBeModerated
	}
	if status >= sender.Status {
		return ErrUserStatusInvalid
	}

	return s.repo.SetStatus(ctx, chatID, m.ID(), int16(status))
}

func (s *Service) SetDevStatus(ctx context.Context, chatID int64, moderator, target chatmember.ChatMember, status permission.Status) error {
	if !s.IsOwner(moderator) {
		return ErrUserCantBeModerated
	}

	return s.repo.SetStatus(ctx, chatID, target.ID(), int16(status))
}

func (s *Service) GetAdminsEnsured(
	ctx context.Context,
	chatID int64,
	sync func(ctx context.Context, chatID int64) (int, error),
) ([]chatmember.ChatMember, error) {
	admins, err := s.chatMemberRepo.ListAdmins(ctx, chatID, permission.StatusModerator)
	if err != nil {
		return nil, err
	}

	if len(admins) > 0 {
		return admins, nil
	}

	if _, err := sync(ctx, chatID); err != nil {
		return nil, err
	}

	return s.chatMemberRepo.ListAdmins(ctx, chatID, permission.StatusModerator)
}

func (s *Service) Kick(ctx context.Context, chatID int64, target, moderator chatmember.ChatMember, reason string) error {
	if !moderator.CanModerate(target) {
		return ErrUserCantBeModerated
	}

	return s.repo.CreateModerationAction(ctx, "kick", chatID, target.ID(), moderator.ID(), reason, time.Time{})
}

func (s *Service) Ban(ctx context.Context, chatID int64, target, moderator chatmember.ChatMember, until time.Time, reason string) error {
	if !moderator.CanModerate(target) {
		return ErrUserCantBeModerated
	}

	return s.repo.CreateModerationAction(ctx, "ban", chatID, target.ID(), moderator.ID(), reason, until)
}

func (s *Service) Mute(ctx context.Context, chatID int64, taget, moderator chatmember.ChatMember, until time.Time, reason string) error {
	if !moderator.CanModerate(taget) {
		return ErrUserCantBeModerated
	}

	if !until.IsZero() {
		now := time.Now()
		duration := until.Sub(now)

		if duration < 30*time.Second || duration > 366*24*time.Hour {
			return ErrInvalidRange
		}
	}

	return s.repo.CreateModerationAction(ctx, "mute", chatID, taget.ID(), moderator.ID(), reason, until)
}

func (s *Service) Warn(
	ctx context.Context,
	ch chat.Chat,
	target,
	moderator chatmember.ChatMember,
	reason string,
	until time.Time,
) (warnsCount int64, err error) {
	chatID := ch.ID

	if !moderator.CanModerate(target) {
		return 0, ErrUserCantBeModerated
	}
	if err := s.repo.CreateModerationAction(ctx, "warn", chatID, target.ID(), moderator.ID(), reason, until); err != nil {
		return 0, err
	}

	count, err := s.repo.GetWarnsCount(ctx, chatID, target.ID())
	if err != nil {
		return 0, err
	}

	maxWarns := ch.MaxWarns
	if int32(count) >= maxWarns {
		_ = s.repo.CreateModerationAction(ctx, "ban", chatID, target.ID(), moderator.ID(), reason, time.Time{})
		_ = s.repo.ClearWarns(ctx, chatID, target.ID())
		return count, nil
	}

	return count, nil
}

func (s *Service) Unban(ctx context.Context, chatID, userID int64) error {
	return s.repo.RemoveModerationActions(ctx, chatID, userID)
}

func (s *Service) Unwarn(ctx context.Context, chatID, userID int64) (int, error) {
	if err := s.repo.RemoveLatestWarn(ctx, chatID, userID); err != nil {
		return 0, err
	}

	count, err := s.repo.GetWarnsCount(ctx, chatID, userID)
	return int(count), err
}

func (s *Service) ClearWarns(ctx context.Context, chatID, userID int64) error {
	return s.repo.ClearWarns(ctx, chatID, userID)
}

func (s *Service) SetMaxWarns(ctx context.Context, chatID int64, maxWarns int) error {
	return s.repo.UpdateChatMaxWarns(ctx, chatID, maxWarns)
}

func (s *Service) GetMaxWarns(ctx context.Context, chatID int64) (int, error) {
	return s.repo.GetChatMaxWarns(ctx, chatID)
}

func (s *Service) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	return s.repo.GetWarnsCount(ctx, chatID, userID)
}

func (s *Service) GetWarns(ctx context.Context, chatID, userID int64) ([]Warn, error) {
	return s.repo.GetActiveWarns(ctx, chatID, userID)
}

func (s *Service) GetWarnsByChat(ctx context.Context, chatID int64) ([]Warn, error) {
	return s.repo.GetActiveWarnsByChat(ctx, chatID)
}

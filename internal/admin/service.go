package admin

import (
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"
)

type ChatMemberStatusProvider interface {
	GetChatMemberStatus(chatID, userID int64) (string, error)
}

type Moderator interface {
	Kick(chatID, userID int64) error
	Ban(chatID, userID int64, untilDate time.Time) error
	Mute(chatID, userID int64, untilDate time.Time) error
	Unban(chatID, userID int64) error
	Unmute(chatID, userID int64) error
}

var (
	ErrUserIsNotAdmin     = errors.New("user is not admin")
	ErrUserIsAlreadyAdmin = errors.New("user is already admin")
	ErrUserIsNotPermitted = errors.New("user is not permitted")
	ErrUserStatusInvalid  = errors.New("user status invalid")
	ErrUserIsCreator      = errors.New("user is creator")
	ErrUserIsProtected    = errors.New("user is protected (admin or creator)")
	ErrInvalidRange       = errors.New("invalid range")
)

type Service struct {
	repo       Repository
	memberRepo member.Repository
	ownerID    int64
}

func NewService(repo Repository, ownerID int64) *Service {
	return &Service{
		repo:    repo,
		ownerID: ownerID,
	}
}

func (s *Service) OwnerID() int64 {
	return s.ownerID
}

func (s *Service) SetStatus(ctx context.Context, sender model.ChatMember, m model.ChatMember, status model.Status) error {
	if !sender.CanModerate(m) {
		return ErrUserIsNotPermitted
	}
	if status >= sender.Status {
		return ErrUserStatusInvalid
	}
	if m.Status == status {
		return ErrUserIsAlreadyAdmin
	}

	return s.repo.SetStatus(ctx, m.ChatID, m.User.ID, int16(status))
}

func (s *Service) SetDevStatus(ctx context.Context, m model.ChatMember, status model.Status) error {
	return s.repo.SetStatus(ctx, m.ChatID, m.User.ID, int16(status))
}

func (s *Service) GetAdminsEnsured(
	ctx context.Context,
	chatID int64,
	sync func(ctx context.Context, chatID int64) (int, error),
) ([]model.ChatMember, error) {
	admins, err := s.repo.GetAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if len(admins) > 0 {
		return admins, nil
	}

	if _, err := sync(ctx, chatID); err != nil {
		return nil, err
	}

	return s.repo.GetAdmins(ctx, chatID)
}

func (s *Service) Kick(ctx context.Context, m model.ChatMember, mod model.ChatMember, reason string) error {
	if !mod.CanModerate(m) {
		return ErrUserIsProtected
	}

	return s.repo.CreateModerationAction(ctx, "kick", m.ChatID, m.User.ID, mod.User.ID, reason, time.Time{})
}

func (s *Service) Ban(ctx context.Context, m model.ChatMember, mod model.ChatMember, until time.Time, reason string) error {
	if !mod.CanModerate(m) {
		return ErrUserIsProtected
	}

	return s.repo.CreateModerationAction(ctx, "ban", m.ChatID, m.User.ID, mod.User.ID, reason, until)
}

func (s *Service) Mute(ctx context.Context, m model.ChatMember, mod model.ChatMember, until time.Time, reason string) error {
	if !mod.CanModerate(m) {
		return ErrUserIsProtected
	}

	if !until.IsZero() {
		now := time.Now()
		duration := until.Sub(now)

		if duration < 30*time.Second || duration > 366*24*time.Hour {
			return ErrInvalidRange
		}
	}

	return s.repo.CreateModerationAction(ctx, "mute", m.ChatID, m.User.ID, mod.User.ID, reason, until)
}

func (s *Service) Warn(ctx context.Context, m model.ChatMember, mod model.ChatMember, reason string, until time.Time) (int, bool, error) {
	if !mod.CanModerate(m) {
		return 0, false, ErrUserIsProtected
	}
	if err := s.repo.CreateModerationAction(ctx, "warn", m.ChatID, m.User.ID, mod.User.ID, reason, until); err != nil {
		return 0, false, err
	}

	count, err := s.repo.GetWarnsCount(ctx, m.ChatID, m.User.ID)
	if err != nil {
		return 0, false, err
	}

	maxWarns, err := s.repo.GetChatMaxWarns(ctx, m.ChatID)
	if err != nil {
		return int(count), false, err
	}

	if int(count) >= maxWarns {
		_ = s.repo.CreateModerationAction(ctx, "ban", m.ChatID, m.User.ID, mod.User.ID, "Превышен лимит предупреждений", time.Time{})
		_ = s.repo.ClearWarns(ctx, m.ChatID, m.User.ID)
		return int(count), true, nil
	}

	return int(count), false, nil
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

func (s *Service) checkCanBeModerated(m model.ChatMember, mod model.ChatMember) error {
	if m.Status >= mod.Status {
		return ErrUserIsProtected
	}
	return nil
}

func (s *Service) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	return s.repo.GetWarnsCount(ctx, chatID, userID)
}

func (s *Service) GetWarns(ctx context.Context, chatID, userID int64) ([]model.Warn, error) {
	return s.repo.GetActiveWarns(ctx, chatID, userID)
}

func (s *Service) GetWarnsByChat(ctx context.Context, chatID int64) ([]model.Warn, error) {
	return s.repo.GetActiveWarnsByChat(ctx, chatID)
}

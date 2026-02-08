package admin

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"log/slog"
	"time"
)

const (
	DevRoleMember  = "member"
	DevRoleAdmin   = "admin"
	DevRoleCreator = "creator"
)

type ChatMemberStatusProvider interface {
	GetChatMemberStatus(chatID, userID int64) (string, error)
}

type Moderator interface {
	Kick(chatID, userID int64) error
	Ban(chatID, userID int64, untilDate *time.Time) error
	Mute(chatID, userID int64, untilDate *time.Time) error
	Unban(chatID, userID int64) error
	Unmute(chatID, userID int64) error
}

var (
	ErrUserIsNotAdmin     = errors.New("user is not admin")
	ErrUserIsAlreadyAdmin = errors.New("user is already admin")
	ErrUserIsCreator      = errors.New("user is creator")
	ErrUserIsProtected    = errors.New("user is protected (admin or creator)")
)

type Service struct {
	repo         Repository
	memberStatus ChatMemberStatusProvider
	moderator    Moderator
}

func NewService(repo Repository, statusProvider ChatMemberStatusProvider, moderator Moderator) *Service {
	return &Service{
		repo:         repo,
		memberStatus: statusProvider,
		moderator:    moderator,
	}
}

func (s *Service) EnsureInitialDeveloper(ctx context.Context, ownerID int64) error {
	count, err := s.repo.GetDevelopersCount(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		slog.Info("No developers found in database, adding root owner as creator", "ownerID", ownerID)
		if err := s.repo.EnsureDeveloperUser(ctx, ownerID); err != nil {
			return err
		}
		return s.repo.SetDeveloperRole(ctx, ownerID, DevRoleCreator)
	}

	return nil
}

func (s *Service) GetDevRole(ctx context.Context, userID int64) (string, error) {
	role, err := s.repo.GetDeveloperRole(ctx, userID)
	if err != nil {
		return DevRoleMember, nil
	}
	return role, nil
}

func (s *Service) SetDevRole(ctx context.Context, userID int64, role string) error {
	return s.repo.SetDeveloperRole(ctx, userID, role)
}

func (s *Service) RemoveDeveloper(ctx context.Context, userID int64) error {
	return s.repo.RemoveDeveloperRole(ctx, userID)
}

func (s *Service) GetAllDevelopers(ctx context.Context) ([]model.User, []string, error) {
	return s.repo.GetAllDevelopers(ctx)
}

func (s *Service) IsDeveloper(ctx context.Context, userID int64) (bool, error) {
	return s.repo.IsDeveloper(ctx, userID)
}

func (s *Service) AddAdmin(ctx context.Context, chatID int64, userID int64) error {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return ErrUserIsAlreadyAdmin
	}

	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if isAdmin {
		return ErrUserIsAlreadyAdmin
	}

	return s.repo.Add(ctx, chatID, userID)
}

func (s *Service) RemoveAdmin(ctx context.Context, chatID int64, userID int64) error {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return ErrUserIsCreator
	}

	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if !isAdmin {
		return ErrUserIsNotAdmin
	}

	return s.repo.Remove(ctx, chatID, userID)
}

func (s *Service) GetAdminsEnsured(
	ctx context.Context,
	chatID int64,
	sync func(ctx context.Context, chatID int64) (int, error),
) ([]model.User, error) {
	admins, err := s.repo.GetFromChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if len(admins) > 0 {
		return admins, nil
	}

	if _, err := sync(ctx, chatID); err != nil {
		return nil, err
	}

	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) IsCreator(ctx context.Context, chatID int64, userID int64) (bool, error) {
	role, _ := s.GetDevRole(ctx, userID)
	if role == DevRoleCreator {
		return true, nil
	}

	isCreator, err := s.repo.IsCreator(ctx, chatID, userID)
	if err == nil {
		return isCreator, nil
	}

	status, err := s.memberStatus.GetChatMemberStatus(chatID, userID)
	if err != nil {
		return false, err
	}

	return status == "creator", nil
}

func (s *Service) GetRole(ctx context.Context, chatID int64, userID int64) (string, error) {
	return s.repo.GetRole(ctx, chatID, userID)
}

func (s *Service) CheckIsAdmin(ctx context.Context, chatID, userID int64) bool {
	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		slog.Error("failed to check admin", "chat_id", chatID, "user_id", userID, "error", err)
		return false
	}
	return isAdmin
}

func (s *Service) CheckIsCreator(ctx context.Context, chatID, userID int64) bool {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		slog.Error("failed to check creator", "chat_id", chatID, "user_id", userID, "error", err)
		return false
	}
	return isCreator
}

func (s *Service) IsAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
	role, _ := s.GetDevRole(ctx, userID)
	if role == DevRoleCreator || role == DevRoleAdmin {
		return true, nil
	}

	isAdmin, err := s.repo.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	status, err := s.memberStatus.GetChatMemberStatus(chatID, userID)
	if err != nil {
		return false, err
	}

	return status == "creator", nil
}

func (s *Service) Kick(ctx context.Context, chatID, userID, modID int64, reason string) error {
	if err := s.checkCanModerate(ctx, chatID, userID); err != nil {
		return err
	}

	if err := s.moderator.Kick(chatID, userID); err != nil {
		return err
	}

	return s.repo.CreateModerationAction(ctx, "kick", chatID, userID, modID, reason, nil)
}

func (s *Service) Ban(ctx context.Context, chatID, userID, modID int64, until *time.Time, reason string) error {
	if err := s.checkCanModerate(ctx, chatID, userID); err != nil {
		return err
	}

	if err := s.moderator.Ban(chatID, userID, until); err != nil {
		return err
	}

	return s.repo.CreateModerationAction(ctx, "ban", chatID, userID, modID, reason, until)
}

func (s *Service) Mute(ctx context.Context, chatID, userID, modID int64, until *time.Time, reason string) error {
	if err := s.checkCanModerate(ctx, chatID, userID); err != nil {
		return err
	}

	if err := s.moderator.Mute(chatID, userID, until); err != nil {
		return err
	}

	return s.repo.CreateModerationAction(ctx, "mute", chatID, userID, modID, reason, until)
}

func (s *Service) Warn(ctx context.Context, chatID, userID, modID int64, reason string, until *time.Time) (int, bool, error) {
	if err := s.checkCanModerate(ctx, chatID, userID); err != nil {
		return 0, false, err
	}

	if err := s.repo.CreateModerationAction(ctx, "warn", chatID, userID, modID, reason, until); err != nil {
		return 0, false, err
	}

	count, err := s.repo.GetWarnsCount(ctx, chatID, userID)
	if err != nil {
		return 0, false, err
	}

	maxWarns, err := s.repo.GetChatMaxWarns(ctx, chatID)
	if err != nil {
		return int(count), false, err
	}

	if int(count) >= maxWarns {
		if err := s.moderator.Ban(chatID, userID, nil); err != nil {
			return int(count), false, err
		}
		_ = s.repo.CreateModerationAction(ctx, "ban", chatID, userID, modID, "Превышен лимит предупреждений", nil)
		_ = s.repo.ClearWarns(ctx, chatID, userID)
		return int(count), true, nil
	}

	return int(count), false, nil
}

func (s *Service) Unban(ctx context.Context, chatID, userID int64) error {
	if err := s.moderator.Unban(chatID, userID); err != nil {
		return err
	}

	return s.repo.RemoveModerationActions(ctx, chatID, userID)
}

func (s *Service) Unmute(_ context.Context, chatID, userID int64) error {
	return s.moderator.Unmute(chatID, userID)
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

func (s *Service) checkCanModerate(ctx context.Context, chatID, userID int64) error {
	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return ErrUserIsProtected
	}

	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return ErrUserIsProtected
	}

	return nil
}

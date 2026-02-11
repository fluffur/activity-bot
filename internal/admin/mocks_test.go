package admin

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type mockRepository struct {
	admins            map[int64]map[int64]bool
	creators          map[int64]map[int64]bool
	developers        map[int64]string
	warns             map[int64]map[int64]int64
	maxWarns          map[int64]int
	moderationActions []mockModerationAction
}

type mockModerationAction struct {
	actionType string
	chatID     int64
	userID     int64
	modID      int64
	reason     string
	until      *time.Time
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		admins:     make(map[int64]map[int64]bool),
		creators:   make(map[int64]map[int64]bool),
		developers: make(map[int64]string),
		warns:      make(map[int64]map[int64]int64),
		maxWarns:   make(map[int64]int),
	}
}

func (m *mockRepository) Add(ctx context.Context, chatID, userID int64) error {
	if m.admins[chatID] == nil {
		m.admins[chatID] = make(map[int64]bool)
	}
	m.admins[chatID][userID] = true
	return nil
}

func (m *mockRepository) Remove(ctx context.Context, chatID, userID int64) error {
	if m.admins[chatID] != nil {
		delete(m.admins[chatID], userID)
	}
	return nil
}

func (m *mockRepository) IsAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
	if m.admins[chatID] == nil {
		return false, nil
	}
	return m.admins[chatID][userID], nil
}

func (m *mockRepository) IsCreator(ctx context.Context, chatID, userID int64) (bool, error) {
	if m.creators[chatID] == nil {
		return false, nil
	}
	return m.creators[chatID][userID], nil
}

func (m *mockRepository) GetFromChat(ctx context.Context, chatID int64) ([]model.User, error) {
	var users []model.User
	if m.admins[chatID] != nil {
		for userID := range m.admins[chatID] {
			users = append(users, model.User{ID: userID})
		}
	}
	return users, nil
}

func (m *mockRepository) GetRole(ctx context.Context, chatID, userID int64) (string, error) {
	if m.creators[chatID] != nil && m.creators[chatID][userID] {
		return "creator", nil
	}
	if m.admins[chatID] != nil && m.admins[chatID][userID] {
		return "administrator", nil
	}
	return "member", nil
}

func (m *mockRepository) CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until *time.Time) error {
	m.moderationActions = append(m.moderationActions, mockModerationAction{
		actionType: actionType,
		chatID:     chatID,
		userID:     userID,
		modID:      modID,
		reason:     reason,
		until:      until,
	})

	if actionType == "warn" {
		if m.warns[chatID] == nil {
			m.warns[chatID] = make(map[int64]int64)
		}
		m.warns[chatID][userID]++
	}

	return nil
}

func (m *mockRepository) RemoveModerationActions(ctx context.Context, chatID, userID int64) error {
	return nil
}

func (m *mockRepository) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	if m.warns[chatID] == nil {
		return 0, nil
	}
	return m.warns[chatID][userID], nil
}

func (m *mockRepository) RemoveLatestWarn(ctx context.Context, chatID, userID int64) error {
	if m.warns[chatID] != nil && m.warns[chatID][userID] > 0 {
		m.warns[chatID][userID]--
	}
	return nil
}

func (m *mockRepository) ClearWarns(ctx context.Context, chatID, userID int64) error {
	if m.warns[chatID] != nil {
		m.warns[chatID][userID] = 0
	}
	return nil
}

func (m *mockRepository) GetChatMaxWarns(ctx context.Context, chatID int64) (int, error) {
	if maxWarns, ok := m.maxWarns[chatID]; ok {
		return maxWarns, nil
	}
	return 3, nil
}

func (m *mockRepository) UpdateChatMaxWarns(ctx context.Context, chatID int64, maxWarns int) error {
	m.maxWarns[chatID] = maxWarns
	return nil
}

func (m *mockRepository) GetDeveloperRole(ctx context.Context, userID int64) (string, error) {
	if role, ok := m.developers[userID]; ok {
		return role, nil
	}
	return "", nil
}

func (m *mockRepository) SetDeveloperRole(ctx context.Context, userID int64, role string) error {
	m.developers[userID] = role
	return nil
}

func (m *mockRepository) RemoveDeveloperRole(ctx context.Context, userID int64) error {
	delete(m.developers, userID)
	return nil
}

func (m *mockRepository) GetDevelopersCount(ctx context.Context) (int64, error) {
	return int64(len(m.developers)), nil
}

func (m *mockRepository) EnsureDeveloperUser(ctx context.Context, userID int64) error {
	return nil
}

func (m *mockRepository) GetAllDevelopers(ctx context.Context) ([]model.User, []string, error) {
	var users []model.User
	var roles []string
	for userID, role := range m.developers {
		users = append(users, model.User{ID: userID})
		roles = append(roles, role)
	}
	return users, roles, nil
}

func (m *mockRepository) IsDeveloper(ctx context.Context, userID int64) (bool, error) {
	_, ok := m.developers[userID]
	return ok, nil
}

type mockChatMemberStatusProvider struct {
	statuses map[int64]map[int64]string
}

func newMockChatMemberStatusProvider() *mockChatMemberStatusProvider {
	return &mockChatMemberStatusProvider{
		statuses: make(map[int64]map[int64]string),
	}
}

func (m *mockChatMemberStatusProvider) GetChatMemberStatus(chatID, userID int64) (string, error) {
	if m.statuses[chatID] == nil {
		return "member", nil
	}
	if status, ok := m.statuses[chatID][userID]; ok {
		return status, nil
	}
	return "member", nil
}

func (m *mockChatMemberStatusProvider) setStatus(chatID, userID int64, status string) {
	if m.statuses[chatID] == nil {
		m.statuses[chatID] = make(map[int64]string)
	}
	m.statuses[chatID][userID] = status
}

type mockModerator struct {
	kicked   map[int64]map[int64]bool
	banned   map[int64]map[int64]bool
	muted    map[int64]map[int64]bool
	unbanned map[int64]map[int64]bool
	unmuted  map[int64]map[int64]bool
}

func newMockModerator() *mockModerator {
	return &mockModerator{
		kicked:   make(map[int64]map[int64]bool),
		banned:   make(map[int64]map[int64]bool),
		muted:    make(map[int64]map[int64]bool),
		unbanned: make(map[int64]map[int64]bool),
		unmuted:  make(map[int64]map[int64]bool),
	}
}

func (m *mockModerator) Kick(chatID, userID int64) error {
	if m.kicked[chatID] == nil {
		m.kicked[chatID] = make(map[int64]bool)
	}
	m.kicked[chatID][userID] = true
	return nil
}

func (m *mockModerator) Ban(chatID, userID int64, untilDate *time.Time) error {
	if m.banned[chatID] == nil {
		m.banned[chatID] = make(map[int64]bool)
	}
	m.banned[chatID][userID] = true
	return nil
}

func (m *mockModerator) Mute(chatID, userID int64, untilDate *time.Time) error {
	if m.muted[chatID] == nil {
		m.muted[chatID] = make(map[int64]bool)
	}
	m.muted[chatID][userID] = true
	return nil
}

func (m *mockModerator) Unban(chatID, userID int64) error {
	if m.unbanned[chatID] == nil {
		m.unbanned[chatID] = make(map[int64]bool)
	}
	m.unbanned[chatID][userID] = true
	return nil
}

func (m *mockModerator) Unmute(chatID, userID int64) error {
	if m.unmuted[chatID] == nil {
		m.unmuted[chatID] = make(map[int64]bool)
	}
	m.unmuted[chatID][userID] = true
	return nil
}

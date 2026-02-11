package member

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type mockMemberRepository struct {
	members     map[int64]map[int64]*model.ChatMember
	leftMembers map[int64]map[int64]bool
}

func newMockMemberRepository() *mockMemberRepository {
	return &mockMemberRepository{
		members:     make(map[int64]map[int64]*model.ChatMember),
		leftMembers: make(map[int64]map[int64]bool),
	}
}

func (m *mockMemberRepository) GetCustomTitle(ctx context.Context, chatID int64, userID int64) (string, error) {
	if m.members[chatID] != nil && m.members[chatID][userID] != nil {
		return m.members[chatID][userID].CustomTitle, nil
	}
	return "", ErrMemberNotFound
}

func (m *mockMemberRepository) UpdateCustomTitle(ctx context.Context, chatID int64, userID int64, title *string) error {
	if m.members[chatID] == nil {
		m.members[chatID] = make(map[int64]*model.ChatMember)
	}
	if m.members[chatID][userID] == nil {
		m.members[chatID][userID] = &model.ChatMember{
			ChatID: chatID,
			User:   model.User{ID: userID},
		}
	}
	if title != nil {
		m.members[chatID][userID].CustomTitle = *title
	} else {
		m.members[chatID][userID].CustomTitle = ""
	}
	return nil
}

func (m *mockMemberRepository) UpdateStatus(ctx context.Context, chatID int64, userID int64, role string) error {
	if m.members[chatID] == nil {
		m.members[chatID] = make(map[int64]*model.ChatMember)
	}
	if m.members[chatID][userID] == nil {
		m.members[chatID][userID] = &model.ChatMember{
			ChatID: chatID,
			User:   model.User{ID: userID},
		}
	}
	m.members[chatID][userID].Status = role
	return nil
}

func (m *mockMemberRepository) FindByChatID(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	var members []model.ChatMember
	if m.members[chatID] != nil {
		for _, member := range m.members[chatID] {
			members = append(members, *member)
		}
	}
	return members, nil
}

func (m *mockMemberRepository) GetWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	var members []model.ChatMember
	if m.members[chatID] != nil {
		for _, member := range m.members[chatID] {
			if member.CustomTitle != "" {
				members = append(members, *member)
			}
		}
	}
	return members, nil
}

func (m *mockMemberRepository) GetAnyWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return m.GetWithCustomTitles(ctx, chatID)
}

func (m *mockMemberRepository) UpsertChatMembers(ctx context.Context, chatID int64, updates []model.ChatMemberUpdate) error {
	if m.members[chatID] == nil {
		m.members[chatID] = make(map[int64]*model.ChatMember)
	}
	for _, update := range updates {
		m.members[chatID][update.User.ID] = &model.ChatMember{
			ChatID: chatID,
			User:   update.User,
			Status: update.Status,
		}
	}
	return nil
}

func (m *mockMemberRepository) MarkLeftNotInList(ctx context.Context, chatID int64, userIDs []int64) error {
	if m.leftMembers[chatID] == nil {
		m.leftMembers[chatID] = make(map[int64]bool)
	}

	activeUsers := make(map[int64]bool)
	for _, uid := range userIDs {
		activeUsers[uid] = true
	}

	if m.members[chatID] != nil {
		for uid := range m.members[chatID] {
			if !activeUsers[uid] {
				m.leftMembers[chatID][uid] = true
			}
		}
	}
	return nil
}

func (m *mockMemberRepository) Get(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	if m.members[chatID] != nil && m.members[chatID][userID] != nil {
		return *m.members[chatID][userID], nil
	}
	return model.ChatMember{}, ErrMemberNotFound
}

func (m *mockMemberRepository) Remove(ctx context.Context, chatID int64, userID int64) error {
	if m.members[chatID] != nil {
		delete(m.members[chatID], userID)
	}
	return nil
}

func (m *mockMemberRepository) EnsureExists(ctx context.Context, chatID int64, userID int64, role string) (model.ChatMember, error) {
	if m.members[chatID] == nil {
		m.members[chatID] = make(map[int64]*model.ChatMember)
	}
	if m.members[chatID][userID] == nil {
		m.members[chatID][userID] = &model.ChatMember{
			ChatID: chatID,
			User:   model.User{ID: userID},
			Status: role,
		}
	}
	return *m.members[chatID][userID], nil
}

func (m *mockMemberRepository) EnsureFull(ctx context.Context, chatID int64, userID int64, role string, firstName, lastName string, username string, weeklyNorm int32) (model.ChatMember, error) {
	return m.EnsureExists(ctx, chatID, userID, role)
}

func (m *mockMemberRepository) SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return nil
}

func (m *mockMemberRepository) SetNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return nil
}

type mockChatRepository struct {
	chats map[int64]*model.Chat
}

func newMockChatRepository() *mockChatRepository {
	return &mockChatRepository{
		chats: make(map[int64]*model.Chat),
	}
}

func (m *mockChatRepository) Ensure(ctx context.Context, chat model.Chat) (model.Chat, error) {
	if m.chats[chat.ID] == nil {
		m.chats[chat.ID] = &chat
	}
	return *m.chats[chat.ID], nil
}

func (m *mockChatRepository) SetNorm(ctx context.Context, chatID int64, norm int32) error {
	return nil
}

func (m *mockChatRepository) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return nil
}

func (m *mockChatRepository) GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error) {
	return int(fallbackNorm), nil
}

func (m *mockChatRepository) GetNewbieThreshold(ctx context.Context, chatID int64) (int, error) {
	return 0, nil
}

func (m *mockChatRepository) GetChat(ctx context.Context, chatID int64) (model.Chat, error) {
	if chat, ok := m.chats[chatID]; ok {
		return *chat, nil
	}
	return model.Chat{}, nil
}

func (m *mockChatRepository) SetChatPrompt(ctx context.Context, chatID int64, prompt string) error {
	return nil
}

func (m *mockChatRepository) SetMaxLadder(ctx context.Context, chatID int64, maxLadder int32) error {
	return nil
}

func (m *mockChatRepository) SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error {
	return nil
}

func (m *mockChatRepository) UpdateCallOnJoin(ctx context.Context, chatID int64, isEnabled bool) error {
	return nil
}

func (m *mockChatRepository) SetWeekStartDay(ctx context.Context, chatID int64, day int) error {
	return nil
}

type mockUserRepository struct {
	users map[int64]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[int64]*model.User),
	}
}

func (m *mockUserRepository) Ensure(ctx context.Context, id int64, username, firstName, lastName string) (model.User, error) {
	if m.users[id] == nil {
		m.users[id] = &model.User{
			ID:        id,
			Username:  &username,
			FirstName: firstName,
			LastName:  lastName,
		}
	}
	return *m.users[id], nil
}

func (m *mockUserRepository) Get(ctx context.Context, id int64) (model.User, error) {
	if user, ok := m.users[id]; ok {
		return *user, nil
	}
	return model.User{}, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	for _, user := range m.users {
		if user.Username != nil && *user.Username == username {
			return *user, nil
		}
	}
	return model.User{}, nil
}

func (m *mockUserRepository) UpsertUsers(ctx context.Context, users []model.User) error {
	for _, u := range users {
		m.users[u.ID] = &u
	}
	return nil
}

func (m *mockUserRepository) GetByCustomTitle(ctx context.Context, chatID int64, title string) (model.User, error) {
	return model.User{}, nil
}

type mockChatAdminsProvider struct {
	admins map[int64][]model.ChatMemberUpdate
}

func newMockChatAdminsProvider() *mockChatAdminsProvider {
	return &mockChatAdminsProvider{
		admins: make(map[int64][]model.ChatMemberUpdate),
	}
}

func (m *mockChatAdminsProvider) GetChatAdmins(chatID int64) ([]model.ChatMemberUpdate, error) {
	if admins, ok := m.admins[chatID]; ok {
		return admins, nil
	}
	return []model.ChatMemberUpdate{}, nil
}

func (m *mockChatAdminsProvider) setAdmins(chatID int64, admins []model.ChatMemberUpdate) {
	m.admins[chatID] = admins
}

var _ Repository = (*mockMemberRepository)(nil)
var _ chat.Repository = (*mockChatRepository)(nil)
var _ user.Repository = (*mockUserRepository)(nil)
var _ ChatAdminsProvider = (*mockChatAdminsProvider)(nil)

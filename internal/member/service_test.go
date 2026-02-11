package member

import (
	"activity-bot/internal/model"
	"context"
	"testing"
)

func TestService_SetMemberTitle(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		userID     int64
		title      *string
		setupMocks func(*mockMemberRepository)
		wantErr    bool
	}{
		{
			name:   "set title successfully",
			chatID: 1,
			userID: 100,
			title:  stringPtr("Admin"),
			setupMocks: func(repo *mockMemberRepository) {
				repo.members[1] = map[int64]*model.ChatMember{
					100: {ChatID: 1, User: model.User{ID: 100}},
				}
			},
			wantErr: false,
		},
		{
			name:       "clear title",
			chatID:     1,
			userID:     100,
			title:      nil,
			setupMocks: func(repo *mockMemberRepository) {},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockMemberRepository()
			chatRepo := newMockChatRepository()
			userRepo := newMockUserRepository()
			adminsProvider := newMockChatAdminsProvider()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, chatRepo, userRepo, adminsProvider, 100)
			err := service.SetMemberTitle(context.Background(), tt.chatID, tt.userID, tt.title)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetMemberTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateChatMembers(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		members    []model.ChatMemberUpdate
		wantErr    bool
		verifyFunc func(*testing.T, *mockMemberRepository, *mockChatRepository, *mockUserRepository)
	}{
		{
			name:   "update members successfully",
			chatID: 1,
			members: []model.ChatMemberUpdate{
				{
					User:   model.User{ID: 100, FirstName: "Alice"},
					Status: "member",
				},
				{
					User:   model.User{ID: 200, FirstName: "Bob"},
					Status: "administrator",
				},
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, memberRepo *mockMemberRepository, chatRepo *mockChatRepository, userRepo *mockUserRepository) {

				if chatRepo.chats[1] == nil {
					t.Error("Chat was not ensured")
				}

				if userRepo.users[100] == nil || userRepo.users[200] == nil {
					t.Error("Users were not upserted")
				}

				if memberRepo.members[1] == nil {
					t.Error("Members were not upserted")
				}
				if len(memberRepo.members[1]) != 2 {
					t.Errorf("Expected 2 members, got %d", len(memberRepo.members[1]))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockMemberRepository()
			chatRepo := newMockChatRepository()
			userRepo := newMockUserRepository()
			adminsProvider := newMockChatAdminsProvider()

			service := NewService(repo, chatRepo, userRepo, adminsProvider, 100)
			err := service.UpdateChatMembers(context.Background(), tt.chatID, tt.members)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateChatMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, repo, chatRepo, userRepo)
			}
		})
	}
}

func TestService_ProcessLeftMember(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		userID     int64
		setupMocks func(*mockMemberRepository)
		wantTitle  string
		wantErr    bool
	}{
		{
			name:   "process left member with custom title",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockMemberRepository) {
				repo.members[1] = map[int64]*model.ChatMember{
					100: {ChatID: 1, User: model.User{ID: 100}, CustomTitle: "VIP"},
				}
			},
			wantTitle: "VIP",
			wantErr:   false,
		},
		{
			name:   "process left member without custom title",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockMemberRepository) {
				repo.members[1] = map[int64]*model.ChatMember{
					100: {ChatID: 1, User: model.User{ID: 100}},
				}
			},
			wantTitle: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockMemberRepository()
			chatRepo := newMockChatRepository()
			userRepo := newMockUserRepository()
			adminsProvider := newMockChatAdminsProvider()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, chatRepo, userRepo, adminsProvider, 100)
			title, err := service.ProcessLeftMember(context.Background(), tt.chatID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessLeftMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if title != tt.wantTitle {
				t.Errorf("ProcessLeftMember() title = %v, want %v", title, tt.wantTitle)
			}

			if !tt.wantErr {
				_, err := repo.Get(context.Background(), tt.chatID, tt.userID)
				if err != ErrMemberNotFound {
					t.Error("Member was not removed")
				}
			}
		})
	}
}

func TestService_SyncChatMembers(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		setupMocks func(*mockChatAdminsProvider, *mockMemberRepository)
		wantCount  int
		wantErr    bool
		verifyFunc func(*testing.T, *mockMemberRepository)
	}{
		{
			name:   "sync members successfully",
			chatID: 1,
			setupMocks: func(provider *mockChatAdminsProvider, repo *mockMemberRepository) {
				provider.setAdmins(1, []model.ChatMemberUpdate{
					{User: model.User{ID: 100}, Status: "creator"},
					{User: model.User{ID: 200}, Status: "administrator"},
					{User: model.User{ID: 300}, Status: "member"},
				})

				repo.members[1] = map[int64]*model.ChatMember{
					999: {ChatID: 1, User: model.User{ID: 999}, Status: "member"},
				}
			},
			wantCount: 3,
			wantErr:   false,
			verifyFunc: func(t *testing.T, repo *mockMemberRepository) {

				if len(repo.members[1]) != 4 {
					t.Errorf("Expected 4 members, got %d", len(repo.members[1]))
				}

				if !repo.leftMembers[1][999] {
					t.Error("Old member was not marked as left")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockMemberRepository()
			chatRepo := newMockChatRepository()
			userRepo := newMockUserRepository()
			adminsProvider := newMockChatAdminsProvider()

			if tt.setupMocks != nil {
				tt.setupMocks(adminsProvider, repo)
			}

			service := NewService(repo, chatRepo, userRepo, adminsProvider, 100)
			count, err := service.SyncChatMembers(context.Background(), tt.chatID)

			if (err != nil) != tt.wantErr {
				t.Errorf("SyncChatMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if count != tt.wantCount {
				t.Errorf("SyncChatMembers() count = %v, want %v", count, tt.wantCount)
			}

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, repo)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

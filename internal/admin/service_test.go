package admin

import (
	"context"
	"testing"
	"time"
)

func TestService_AddAdmin(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		userID        int64
		setupMocks    func(*mockRepository, *mockChatMemberStatusProvider)
		wantErr       bool
		expectedError error
	}{
		{
			name:   "successfully add admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				// User is not admin or creator
			},
			wantErr: false,
		},
		{
			name:   "fail when user is already admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.admins[1] = map[int64]bool{100: true}
			},
			wantErr:       true,
			expectedError: ErrUserIsAlreadyAdmin,
		},
		{
			name:   "fail when user is creator",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.creators[1] = map[int64]bool{100: true}
			},
			wantErr:       true,
			expectedError: ErrUserIsAlreadyAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo, statusProvider)
			}

			service := NewService(repo, statusProvider, moderator)
			err := service.AddAdmin(context.Background(), tt.chatID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("AddAdmin() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.wantErr {
				// Verify user was added
				isAdmin, _ := repo.IsAdmin(context.Background(), tt.chatID, tt.userID)
				if !isAdmin {
					t.Error("User was not added as admin")
				}
			}
		})
	}
}

func TestService_RemoveAdmin(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		userID        int64
		setupMocks    func(*mockRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:   "successfully remove admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository) {
				repo.admins[1] = map[int64]bool{100: true}
			},
			wantErr: false,
		},
		{
			name:   "fail when user is not admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository) {
				// User is not admin
			},
			wantErr:       true,
			expectedError: ErrUserIsNotAdmin,
		},
		{
			name:   "fail when user is creator",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository) {
				repo.creators[1] = map[int64]bool{100: true}
			},
			wantErr:       true,
			expectedError: ErrUserIsCreator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, statusProvider, moderator)
			err := service.RemoveAdmin(context.Background(), tt.chatID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("RemoveAdmin() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}

func TestService_IsAdmin(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		userID     int64
		setupMocks func(*mockRepository, *mockChatMemberStatusProvider)
		want       bool
	}{
		{
			name:   "developer with creator role is admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.developers[100] = DevRoleCreator
			},
			want: true,
		},
		{
			name:   "developer with admin role is admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.developers[100] = DevRoleAdmin
			},
			want: true,
		},
		{
			name:   "user in admin table is admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.admins[1] = map[int64]bool{100: true}
			},
			want: true,
		},
		{
			name:   "telegram creator is admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				status.setStatus(1, 100, "creator")
			},
			want: true,
		},
		{
			name:   "regular member is not admin",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				// No special status
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo, statusProvider)
			}

			service := NewService(repo, statusProvider, moderator)
			got, err := service.IsAdmin(context.Background(), tt.chatID, tt.userID)

			if err != nil {
				t.Errorf("IsAdmin() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Kick(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		userID        int64
		modID         int64
		reason        string
		setupMocks    func(*mockRepository, *mockChatMemberStatusProvider)
		wantErr       bool
		expectedError error
	}{
		{
			name:   "successfully kick regular user",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				// Regular user
			},
			wantErr: false,
		},
		{
			name:   "fail to kick admin",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.admins[1] = map[int64]bool{100: true}
			},
			wantErr:       true,
			expectedError: ErrUserIsProtected,
		},
		{
			name:   "fail to kick creator",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository, status *mockChatMemberStatusProvider) {
				repo.creators[1] = map[int64]bool{100: true}
			},
			wantErr:       true,
			expectedError: ErrUserIsProtected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo, statusProvider)
			}

			service := NewService(repo, statusProvider, moderator)
			err := service.Kick(context.Background(), tt.chatID, tt.userID, tt.modID, tt.reason)

			if (err != nil) != tt.wantErr {
				t.Errorf("Kick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("Kick() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.wantErr {
				// Verify kick was called
				if !moderator.kicked[tt.chatID][tt.userID] {
					t.Error("Moderator.Kick was not called")
				}
			}
		})
	}
}

func TestService_Warn(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		userID     int64
		modID      int64
		reason     string
		setupMocks func(*mockRepository)
		wantCount  int
		wantBanned bool
		wantErr    bool
	}{
		{
			name:   "first warn",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository) {
				repo.maxWarns[1] = 3
			},
			wantCount:  1,
			wantBanned: false,
			wantErr:    false,
		},
		{
			name:   "warn triggers ban at limit",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository) {
				repo.maxWarns[1] = 3
				repo.warns[1] = map[int64]int64{100: 2} // Already has 2 warns
			},
			wantCount:  3,
			wantBanned: true,
			wantErr:    false,
		},
		{
			name:   "second warn does not trigger ban",
			chatID: 1,
			userID: 100,
			modID:  200,
			reason: "spam",
			setupMocks: func(repo *mockRepository) {
				repo.maxWarns[1] = 5
				repo.warns[1] = map[int64]int64{100: 1}
			},
			wantCount:  2,
			wantBanned: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, statusProvider, moderator)
			until := time.Now().Add(24 * time.Hour)
			count, banned, err := service.Warn(context.Background(), tt.chatID, tt.userID, tt.modID, tt.reason, &until)

			if (err != nil) != tt.wantErr {
				t.Errorf("Warn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if count != tt.wantCount {
				t.Errorf("Warn() count = %v, want %v", count, tt.wantCount)
			}

			if banned != tt.wantBanned {
				t.Errorf("Warn() banned = %v, want %v", banned, tt.wantBanned)
			}

			if tt.wantBanned && !moderator.banned[tt.chatID][tt.userID] {
				t.Error("User should have been banned but wasn't")
			}
		})
	}
}

func TestService_Unwarn(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		userID     int64
		setupMocks func(*mockRepository)
		wantCount  int
	}{
		{
			name:   "unwarn reduces count",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository) {
				repo.warns[1] = map[int64]int64{100: 3}
			},
			wantCount: 2,
		},
		{
			name:   "unwarn from zero stays at zero",
			chatID: 1,
			userID: 100,
			setupMocks: func(repo *mockRepository) {
				// No warns
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, statusProvider, moderator)
			count, err := service.Unwarn(context.Background(), tt.chatID, tt.userID)

			if err != nil {
				t.Errorf("Unwarn() error = %v", err)
				return
			}

			if count != tt.wantCount {
				t.Errorf("Unwarn() count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestService_EnsureInitialDeveloper(t *testing.T) {
	tests := []struct {
		name       string
		ownerID    int64
		setupMocks func(*mockRepository)
		wantRole   string
	}{
		{
			name:    "add owner as creator when no developers exist",
			ownerID: 100,
			setupMocks: func(repo *mockRepository) {
				// No developers
			},
			wantRole: DevRoleCreator,
		},
		{
			name:    "do nothing when developers already exist",
			ownerID: 100,
			setupMocks: func(repo *mockRepository) {
				repo.developers[200] = DevRoleCreator
			},
			wantRole: "", // Owner should not be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			statusProvider := newMockChatMemberStatusProvider()
			moderator := newMockModerator()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewService(repo, statusProvider, moderator)
			err := service.EnsureInitialDeveloper(context.Background(), tt.ownerID)

			if err != nil {
				t.Errorf("EnsureInitialDeveloper() error = %v", err)
				return
			}

			if tt.wantRole != "" {
				role, _ := repo.GetDeveloperRole(context.Background(), tt.ownerID)
				if role != tt.wantRole {
					t.Errorf("Owner role = %v, want %v", role, tt.wantRole)
				}
			}
		})
	}
}

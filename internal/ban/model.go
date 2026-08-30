package ban

import "time"

type UserBan struct {
	UserID    int64
	Reason    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type ChatBan struct {
	ChatID    int64
	Reason    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

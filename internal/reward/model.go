package reward

import "time"

type Reward struct {
	ID        int64
	ChatID    int64
	UserID    int64
	AuthorID  int64
	Rank      int16
	Reason    string
	CreatedAt time.Time
}

func NewReward(chatID, userID, authorID int64, rank int16, reason string) Reward {
	return Reward{
		ChatID:    chatID,
		UserID:    userID,
		AuthorID:  authorID,
		Rank:      rank,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
}

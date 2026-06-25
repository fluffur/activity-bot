package repository

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/norm"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
)

func mapList[T any, V any](list []T, fn func(T) V) []V {
	l := make([]V, len(list))
	for i, v := range list {
		l[i] = fn(v)
	}
	return l
}

func mapUser(u db.User) user.User {
	return user.User{
		ID:        u.ID,
		FirstName: u.FirstName.String,
		LastName:  u.LastName.String,
		Username:  u.Username.String,
		Gender:    user.Gender(u.Gender),
		Emojis:    u.Emoji.String,
		IsBot:     u.IsBot,
		CreatedAt: u.CreatedAt.Time,
	}
}

func mapChat(c db.Chat) chat.Chat {
	return chat.Chat{
		ID:                   c.ID,
		Title:                c.Title,
		NormWarn:             c.NormWarn.Int32,
		NormBan:              c.NormBan.Int32,
		NewbieThresholdDays:  c.NewbieThresholdDays,
		AISystemPrompt:       c.AiSystemPrompt.String,
		MaxWarns:             c.MaxWarns,
		MaxLadder:            c.MaxLadder,
		WelcomeCallMessage:   c.WelcomeCallMessage.String,
		CallOnJoin:           c.CallOnJoin,
		SkipCallConfirmation: c.SkipCallConfirmation,
		WeekStartDay:         c.WeekStartDay,
		CommandPrefix:        c.CommandPrefix.String,
		AllowPrefixless:      c.AllowPrefixless,
		MentionsPerMessage:   c.MentionsPerMessage,
		MentionTypes:         chat.MentionTypes(c.MentionTypes),
		TagsEnabled:          c.TagsEnabled,
		WeekStartTimeMicros:  c.WeekStartTime.Microseconds,
		EmojisEnabled:        c.EmojisEnabled,
	}
}

func mapChatMember(m db.ChatMember) chatmember.ChatMember {
	return chatmember.ChatMember{
		User: user.User{
			ID: m.UserID,
		},
		Chat: chat.Chat{
			ID: m.ChatID,
		},
		RestUntil:       m.RestUntil.Time,
		RestReason:      m.RestReason.String,
		Tag:             m.Tag.String,
		Status:          chatmember.Status(m.Status),
		Emojis:          m.Emoji.String,
		JoinedAt:        m.JoinedAt.Time,
		LeftAt:          m.LeftAt.Time,
		ExcludeFromCall: m.ExcludeFromCall,
	}
}

func mapChatMemberFull(m db.ChatMember, c db.Chat, u db.User) chatmember.ChatMember {
	cm := mapChatMember(m)

	cm.Chat = mapChat(c)
	cm.User = mapUser(u)

	return cm
}

func mapNorm(n db.ChatNorm) norm.Norm {
	return norm.Norm{
		ID:     n.ID,
		ChatID: n.ChatID,
		Name:   n.Name,
		Value:  n.Value,
	}
}

func mapChatStats(msgCount int64, cm db.ChatMember, u db.User) stats.ChatStats {
	return stats.ChatStats{
		ChatMember:    mapChatMemberFull(cm, db.Chat{}, u),
		MessagesCount: msgCount,
	}
}

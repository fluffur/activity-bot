package postgres

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/norm"
	"activity-bot/internal/permission"
	"activity-bot/internal/rest"
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
		NewbieThresholdDays:  c.NewbieThresholdDays,
		AISystemPrompt:       c.AiSystemPrompt.String,
		MaxLadder:            c.MaxLadder,
		WelcomeCallMessage:   c.WelcomeCallMessage.String,
		CallOnJoin:           c.CallOnJoin,
		SkipCallConfirmation: c.SkipCallConfirmation,
		WeekStartDay:         c.WeekStartDay,
		WeekStartTime:        c.WeekStartTime.Microseconds,
		CommandPrefix:        c.CommandPrefix.String,
		AllowPrefixless:      c.AllowPrefixless,
		MentionsPerMessage:   c.MentionsPerMessage,
		MentionTypes:         chat.MentionTypes(c.MentionTypes),
		TagsEnabled:          c.TagsEnabled,
		WeekStartTimeMicros:  c.WeekStartTime.Microseconds,
		MaxWarns:             c.MaxWarns,
		EmojisEnabled:        c.EmojisEnabled,
		AllowPolygamy:        c.AllowPolygamy,
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
		Status:          permission.Status(m.Status),
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

func mapProfileStats(s db.UserStatsRow) stats.ProfileStats {
	return stats.ProfileStats{
		ChatMember:        mapChatMemberFull(s.ChatMember, db.Chat{}, s.User),
		DayCount:          s.DayCount,
		DayRollingCount:   s.DayRollingCount,
		WeekCount:         s.WeekCount,
		WeekRollingCount:  s.WeekRollingCount,
		MonthCount:        s.MonthCount,
		MonthRollingCount: s.MonthRollingCount,
		AllTimeCount:      s.AllTimeCount,
	}
}

func mapRestRequest(rr db.RestRequest) rest.Request {
	return rest.Request{
		ID:          rr.ID.Int64,
		ChatID:      rr.ChatID,
		UserID:      rr.UserID,
		RequestedAt: rr.RequestedAt.Time,
		UpdatedAt:   rr.UpdatedAt.Time,
		RestUntil:   rr.RestUntil.Time,
		Status:      string(rr.Status),
		MessageID:   rr.MessageID.Int64,
		Reason:      rr.Reason.String,
	}
}

func mapRestRequestFull(rr db.RestRequest, cm db.ChatMember, u db.User) rest.Request {
	r := mapRestRequest(rr)

	r.ChatMember = mapChatMemberFull(cm, db.Chat{}, u)

	return r
}

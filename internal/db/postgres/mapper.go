package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/norm"
)

func mapUser(u db.User) model.User {
	return model.User{
		ID:            u.ID,
		FirstName:     u.FirstName.String,
		LastName:      u.LastName.String,
		Username:      u.Username.String,
		Gender:        u.Gender,
		Emoji:         u.Emoji.String,
		CustomEmojiID: u.CustomEmojiID.String,
	}
}

func mapChat(c db.Chat) model.Chat {
	return model.Chat{
		ID:                  c.ID,
		Title:               c.Title,
		NormWarn:            c.NormWarn.Int32,
		NormBan:             c.NormBan.Int32,
		NewbieThresholdDays: c.NewbieThresholdDays,
		AISystemPrompt:      c.AiSystemPrompt.String,
		MaxWarns:            c.MaxWarns,
		MaxLadder:           c.MaxLadder,
		WelcomeCallMessage:  c.WelcomeCallMessage.String,
		CallOnJoin:          c.CallOnJoin,
		WeekStartDay:        c.WeekStartDay,
		CommandPrefix:       c.CommandPrefix.String,
		AllowPrefixless:     c.AllowPrefixless,
		MentionsPerMessage:  c.MentionsPerMessage,
		MentionTypes:        c.MentionTypes,
		TagsEnabled:         c.TagsEnabled,
		WeekStartTime:       helpers.MicrosecondsToTime(c.WeekStartTime.Microseconds),
	}
}

func mapChatFromRow(c db.EnsureChatExistsRow) model.Chat {
	return mapChat(db.Chat(c))
}

func mapChatMember(m db.ChatMember) model.ChatMember {
	return model.ChatMember{
		User: model.User{
			ID: m.UserID,
		},
		ChatID:     m.ChatID,
		RestUntil:  m.RestUntil.Time,
		RestReason: m.RestReason.String,
		Tag:        m.Tag.String,
		Status:     model.Status(m.Status),
		Emoji:      m.Emoji.String,
		Emojis:     m.EmojiJson,
		JoinedAt:   m.JoinedAt.Time,
		LeftAt:     m.LeftAt.Time,
	}
}

func mapChatMemberFull(m db.ChatMember, u db.User) model.ChatMember {
	res := mapChatMember(m)
	res.User = mapUser(u)
	return res
}

func mapMembers[T any](members []T, mapper func(T) model.ChatMember) []model.ChatMember {
	return mapMany(members, mapper)
}

func mapMessageReportRow(m db.ChatMemberMessageStatsByChatRow) model.ChatMemberMessageCount {
	return model.ChatMemberMessageCount{
		Chat:         mapChat(m.Chat),
		ChatMember:   mapChatMemberFull(m.ChatMember, m.User),
		MessageCount: m.MessagesCount,
	}
}

func mapMessageReportOneRow(m db.ChatMemberMessageStatsByUserRow) model.ChatMemberStats {
	return model.ChatMemberStats{
		ChatMember:        mapChatMemberFull(m.ChatMember, m.User),
		Chat:              mapChat(m.Chat),
		DayCount:          m.DayCount,
		DayRollingCount:   m.DayRollingCount,
		WeekCount:         m.WeekCount,
		WeekRollingCount:  m.WeekRollingCount,
		MonthCount:        m.MonthCount,
		MonthRollingCount: m.MonthRollingCount,
		AllTime:           m.AllTimeCount,
	}
}

func mapInactiveChatMembersRow(m db.InactiveChatMembersRow) model.ChatMember {
	res := mapUser(m.User)
	return model.ChatMember{
		User:      res,
		Tag:       m.Tag.String,
		Status:    model.Status(m.Status),
		RestUntil: m.RestUntil.Time,
	}
}

func mapMessageActivity(row db.ChatMessageActivityDailyRow) model.MessageActivity {
	return model.MessageActivity{
		Count: row.MessagesCount,
		Date:  row.Day.Time,
	}
}

func mapRestRequest(er db.RestRequest) model.RestRequest {
	return model.RestRequest{
		ID:          er.ID.Int64,
		ChatID:      er.ChatID,
		UserID:      er.UserID,
		RequestedAt: er.RequestedAt.Time,
		UpdatedAt:   er.UpdatedAt.Time,
		RestUntil:   er.RestUntil.Time,
		Status:      string(er.Status),
		MessageID:   er.MessageID.Int64,
		Reason:      er.Reason.String,
	}
}

func mapApprovedRestRequest(rr db.RestRequest, cm db.ChatMember, u db.User) model.ApprovedRestRequest {
	return model.ApprovedRestRequest{
		RestRequest: mapRestRequest(rr),
		ChatMember:  mapChatMemberFull(cm, u),
	}
}

func mapChatNorm(n db.ChatNorm) norm.ChatNorm {
	return norm.ChatNorm{
		Name:  n.Name,
		Value: n.Value,
	}
}

func mapMany[T any, V any](input []T, mapper func(T) V) []V {
	result := make([]V, len(input))
	for i, v := range input {
		result[i] = mapper(v)
	}
	return result
}

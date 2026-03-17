package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
)

// User mapping helpers
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

// Chat mapping helpers
func mapChat(c db.Chat) model.Chat {
	return model.Chat{
		ID:                  c.ID,
		Title:               c.Title,
		NormWarn:            c.NormWarn.Int32,
		NormBan:             c.NormBan.Int32,
		NewbieThresholdDays: c.NewbieThresholdDays,
		AISystemPrompt:      c.AiSystemPrompt.String,
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

func mapChats(chats []db.Chat) []model.Chat {
	result := make([]model.Chat, len(chats))
	for i, c := range chats {
		result[i] = mapChat(c)
	}
	return result
}

// ChatMember mapping helpers
func mapChatMember(m db.ChatMember) model.ChatMember {
	return model.ChatMember{
		ChatID: m.ChatID,
		User: model.User{
			ID: m.UserID,
		},
		RestUntil:   m.RestUntil.Time,
		RestReason:  m.RestReason.String,
		CustomTitle: m.CustomTitle.String,
		Status:      m.Status,
	}
}

func mapChatMemberFull(m db.ChatMember, u db.User) model.ChatMember {
	res := mapChatMember(m)
	res.User = mapUser(u)
	return res
}

func mapMembers[T any](members []T, mapper func(T) model.ChatMember) []model.ChatMember {
	result := make([]model.ChatMember, len(members))
	for i, m := range members {
		result[i] = mapper(m)
	}
	return result
}

func mapMessageReportRow(m db.MessageReportRow) model.MessageReportMember {
	return model.MessageReportMember{
		User:                mapUser(m.User),
		MessagesCount:       int(m.MessagesCount),
		NormWarn:            int(m.NormWarn.Int32),
		NormBan:             int(m.NormBan.Int32),
		NewbieThresholdDays: int(m.NewbieThresholdDays),
		Status:              m.ChatMember.Status,
		CustomTitle:         m.ChatMember.CustomTitle.String,
		JoinedAt:            m.ChatMember.JoinedAt.Time,
	}
}

func mapMessageReportOneRow(m db.MessageReportOneRow) model.MemberStats {
	return model.MemberStats{
		User:              mapUser(m.User),
		DayCount:          int(m.DayCount),
		DayRollingCount:   int(m.DayRollingCount),
		WeekCount:         int(m.WeekCount),
		WeekRollingCount:  int(m.WeekRollingCount),
		MonthCount:        int(m.MonthCount),
		MonthRollingCount: int(m.MonthRollingCount),
		AllTime:           int(m.AllTimeCount),
		NormBan:           int(m.NormBan.Int32),
		NormWarn:          int(m.NormWarn.Int32),
		JoinedAt:          m.ChatMember.JoinedAt.Time,
		RestUntil:         m.ChatMember.RestUntil.Time,
		NewbieThreshold:   int(m.NewbieThresholdDays),
		Status:            m.ChatMember.Status,
		CustomTitle:       m.ChatMember.CustomTitle.String,
		LeftAt:            m.ChatMember.LeftAt.Time,
	}
}

func mapInactiveChatMembersRow(m db.InactiveChatMembersRow) model.ChatMember {
	res := mapUser(m.User)
	return model.ChatMember{
		User:        res,
		CustomTitle: m.CustomTitle.String,
		Status:      m.Status,
		RestUntil:   m.RestUntil.Time,
	}
}

func mapMessageActivity(row db.MessageActivityByDayRow) model.MessageActivity {
	return model.MessageActivity{
		Count: row.MessagesCount,
		Date:  row.Day.Time,
	}
}

func mapRestRequest(er db.RestRequest) model.RestRequest {
	return model.RestRequest{
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

func mapApprovedRestRequests[T any](rows []T, mapper func(T) model.ApprovedRestRequest) []model.ApprovedRestRequest {
	result := make([]model.ApprovedRestRequest, len(rows))
	for i, row := range rows {
		result[i] = mapper(row)
	}
	return result
}

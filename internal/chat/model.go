package chat

import "time"

type Chat struct {
	ID                   int64
	Title                string
	NormWarn             int32
	NormBan              int32
	NewbieThresholdDays  int32
	AISystemPrompt       string
	MaxLadder            int32
	WelcomeCallMessage   string
	CallOnJoin           bool
	SkipCallConfirmation bool
	WeekStartDay         int16
	CommandPrefix        string
	AllowPrefixless      bool
	MentionsPerMessage   int32
	MentionTypes         int32
	TagsEnabled          bool
	WeekStartTime        int64
	MaxWarns             int32
	EmojisEnabled        bool
	RemovedAt            time.Time
}

func New(id int64, title string) Chat {
	return Chat{
		ID:                   id,
		Title:                title,
		NewbieThresholdDays:  3,
		MaxLadder:            0,
		CallOnJoin:           false,
		WeekStartDay:         1,
		MaxWarns:             3,
		AllowPrefixless:      true,
		MentionsPerMessage:   5,
		MentionTypes:         0,
		TagsEnabled:          true,
		WeekStartTime:        0,
		EmojisEnabled:        true,
		SkipCallConfirmation: false,
	}
}

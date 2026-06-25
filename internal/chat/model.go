package chat

import "time"

type MentionTypes int32

const (
	MentionEmoji MentionTypes = 1 << iota
	MentionName
	MentionRole
)

func (t *MentionTypes) Has(flag MentionTypes) bool {
	if t == nil {
		return false
	}
	return *t&flag == flag
}

func (t *MentionTypes) Add(flag MentionTypes) {
	*t |= flag
}

func (t *MentionTypes) Remove(flag MentionTypes) {
	*t &^= flag
}

type Chat struct {
	ID                   int64
	Title                string
	Lang                 string
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
	MentionTypes         MentionTypes
	TagsEnabled          bool
	WeekStartTimeMicros  int64
	MaxWarns             int32
	EmojisEnabled        bool
	RemovedAt            time.Time
}

func New(id int64, title string) Chat {
	return Chat{
		ID:                   id,
		Title:                title,
		Lang:                 "ru",
		NormWarn:             0,
		NormBan:              0,
		NewbieThresholdDays:  3,
		AISystemPrompt:       "",
		MaxLadder:            0,
		WelcomeCallMessage:   "",
		CallOnJoin:           false,
		SkipCallConfirmation: false,
		WeekStartDay:         1,
		CommandPrefix:        "",
		AllowPrefixless:      true,
		MentionsPerMessage:   5,
		MentionTypes:         0,
		TagsEnabled:          true,
		WeekStartTimeMicros:  0,
		MaxWarns:             3,
		EmojisEnabled:        true,
		RemovedAt:            time.Time{},
	}
}

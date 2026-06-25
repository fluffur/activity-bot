package stats

import "activity-bot/internal/chatmember"

type UserNormStat struct {
	Norm     string
	Required int32
	Actual   int64
	Passed   bool
}

type UserStats struct {
	ChatMember chatmember.ChatMember
	Norms      []UserNormStat
}

type UserResult struct {
	Member   chatmember.ChatMember
	Messages int64
}

type NormResult struct {
	NormName string
	Required int32

	Passed []UserResult
	Failed []UserResult
}

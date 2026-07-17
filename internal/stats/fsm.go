package stats

import (
	"time"
)

type State string

const (
	StateIdle                    State = ""
	StateAwaitNorm               State = "await_norm"
	StateAwaitSummonText         State = "await_summon_text"
	StateAwaitInactiveSummonText State = "await_summon_inactive_text"
)

type StateData struct {
	ChatID     int64
	UserID     int64
	SummonText string
	NormName   string
	FromDate   time.Time
	ToDate     time.Time
}

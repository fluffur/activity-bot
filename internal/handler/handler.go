package handler

import "activity-bot/internal/command"

type Handler interface {
	Actions() []*command.ActionDef
}

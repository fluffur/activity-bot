package command

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
)

type TriggerType string

const (
	TriggerCommand  TriggerType = "command"
	TriggerCallback TriggerType = "callback"
)

type Category string

type Scope int

const (
	ScopeAny     Scope = 0
	ScopePrivate Scope = 1
	ScopeGroup   Scope = 2
)

type ActionDef struct {
	Key string

	Aliases []string

	Trigger TriggerType

	Parent *ActionDef

	MinStatus   permission.Status
	Category    Category
	Description i18n.MessageID
	Examples    []i18n.MessageID
	Scope       Scope

	ShowInHelp bool
	Rules      []rule.Rule
}

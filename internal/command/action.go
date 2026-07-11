package command

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"

	"github.com/gotd/botapi"
)

type Category string

type Scope int

const (
	ScopeAny     Scope = 0
	ScopePrivate Scope = 1
	ScopeGroup   Scope = 2
)

func (s Scope) BelongsTo(scope Scope) bool {
	return scope == s || scope == ScopeAny
}

type Trigger interface {
	isTrigger()
	IndexKeys() []string
}

type CommandTrigger struct {
	Aliases  []string
	Scope    Scope
	Rules    []rule.Rule
	Prefixes []string
}

func (c *CommandTrigger) isTrigger() {}

func (c *CommandTrigger) IndexKeys() []string {
	return c.Aliases
}

type CallbackTrigger struct {
	Data   string
	Prefix bool
}

func (*CallbackTrigger) isTrigger() {}

func (*CallbackTrigger) IndexKeys() []string {
	return nil
}

type Action struct {
	Key     string
	Handler botapi.Handler

	Trigger Trigger

	Permission             permission.Status
	IgnorePermissionDenied bool

	Category    Category
	Description i18n.MessageID
	Examples    []i18n.MessageID

	ShowInHelp bool
	AllowDev   bool

	ExtraPredicates []botapi.Predicate
}

func (a *Action) CallbackData() (string, bool) {
	t, ok := a.Trigger.(*CallbackTrigger)
	if !ok {
		return "", false
	}

	return t.Data, true
}

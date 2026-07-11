package option

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"

	"github.com/gotd/botapi"
)

type Option func(*command.Action)

func WithRules(rules ...rule.Rule) Option {
	return func(c *command.Action) {
		c.Trigger.(*command.CommandTrigger).Rules = rules
	}
}

func WithExamples(ex ...i18n.MessageID) Option {
	return func(c *command.Action) {
		c.Examples = ex
	}
}

func WithAliases(aliases ...string) Option {
	return func(c *command.Action) {
		c.Trigger.(*command.CommandTrigger).Aliases = aliases
	}
}

func WithScope(scope command.Scope) Option {
	return func(c *command.Action) {
		c.Trigger.(*command.CommandTrigger).Scope = scope
	}
}

func WithPredicates(predicates ...botapi.Predicate) Option {
	return func(c *command.Action) {
		c.ExtraPredicates = predicates
	}
}

func WithPermission(p permission.Status) Option {
	return func(c *command.Action) {
		c.Permission = p
	}
}

func IgnorePermissionCheck() Option {
	return func(c *command.Action) {
		c.IgnorePermissionDenied = true
	}
}

func Hidden() Option {
	return func(c *command.Action) {
		c.ShowInHelp = false
	}
}

func AllowDev() Option {
	return func(c *command.Action) {
		c.AllowDev = true
	}
}

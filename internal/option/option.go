package option

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/rule"
)

type Option func(*command.ActionDef)

func WithRules(rules ...rule.Rule) Option {
	return func(c *command.ActionDef) {
		c.Rules = rules
	}
}

func WithExamples(ex ...i18n.MessageID) Option {
	return func(c *command.ActionDef) {
		c.Examples = ex
	}
}

func WithAliases(aliases ...string) Option {
	return func(c *command.ActionDef) {
		c.Aliases = aliases
	}
}

func WithScope(scope command.Scope) Option {
	return func(c *command.ActionDef) {
		c.Scope = scope
	}
}

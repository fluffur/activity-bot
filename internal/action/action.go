package action

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
)

func NewCommand(
	key string,
	description i18n.MessageID,
	category command.Category,
	minStatus permission.Status,
	opts ...option.Option,
) *command.ActionDef {
	cmd := &command.ActionDef{
		Key:         key,
		Trigger:     command.TriggerCommand,
		MinStatus:   minStatus,
		Category:    category,
		Description: description,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func NewCallback(
	key string,
	category command.Category,
	minStatus permission.Status,
	parent *command.ActionDef,
	opts ...option.Option,
) *command.ActionDef {
	cmd := &command.ActionDef{
		Key:        key,
		Trigger:    command.TriggerCallback,
		Category:   category,
		MinStatus:  minStatus,
		Parent:     parent,
		ShowInHelp: false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

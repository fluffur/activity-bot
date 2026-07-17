package action

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"

	"github.com/gotd/botapi"
)

var defaultPrefixes = []string{"!", "/", ".", "фм", "крыс"}

func NewCommand(
	key string,
	handler botapi.Handler,
	description i18n.MessageID,
	category command.Category,
	opts ...option.Option,
) *command.Action {
	cmd := &command.Action{
		Key:     key,
		Handler: handler,
		Trigger: &command.CommandTrigger{
			Scope:    command.ScopeGroup,
			Prefixes: defaultPrefixes,
		},
		Permission:  permission.StatusMember,
		Category:    category,
		Description: description,
		ShowInHelp:  true,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func NewCallback(
	key,
	data string,
	handler botapi.Handler,
	category command.Category,
	opts ...option.Option,
) *command.Action {
	cmd := &command.Action{
		Key:     key,
		Handler: handler,
		Trigger: &command.CallbackTrigger{
			Data: data,
		},
		Category:   category,
		Permission: permission.StatusMember,
		ShowInHelp: false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func NewCallbackPrefix(
	key,
	data string,
	handler botapi.Handler,
	category command.Category,
	opts ...option.Option,
) *command.Action {
	cmd := &command.Action{
		Key:     key,
		Handler: handler,
		Trigger: &command.CallbackTrigger{
			Data:   data,
			Prefix: true,
		},
		Category:   category,
		Permission: permission.StatusMember,
		ShowInHelp: false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func NewRPCommand(
	key string,
	handler botapi.Handler,
	category command.Category,
	opts ...option.Option,
) *command.Action {
	cmd := &command.Action{
		Key:     key,
		Handler: handler,
		Trigger: &command.RPTrigger{
			Scope:    command.ScopeGroup,
			Prefixes: defaultPrefixes,
		},
		Permission: permission.StatusMember,
		Category:   category,
		ShowInHelp: false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func NewMessage(
	handler botapi.Handler,
	category command.Category,
	opts ...option.Option,
) *command.Action {
	cmd := &command.Action{
		Key:        "message",
		Handler:    handler,
		Trigger:    &command.MessageTrigger{},
		Category:   category,
		Permission: permission.StatusMember,
		ShowInHelp: false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

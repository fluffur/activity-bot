package register

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"

	"github.com/gotd/botapi"
)

func Attach(
	bot *botapi.Bot,
	registry *command.Registry,
	permissions *predicate.PermissionChecker,
	rules *predicate.RuleChecker,
) {
	for _, action := range registry.All() {
		registerAction(bot, action, permissions, rules)
	}
}

func registerAction(
	bot *botapi.Bot,
	action *command.Action,
	permissions *predicate.PermissionChecker,
	rules *predicate.RuleChecker,
) {
	switch t := action.Trigger.(type) {
	case *command.CommandTrigger:
		bot.OnMessage(
			action.Handler,
			buildCommandPredicates(action, t, permissions, rules)...,
		)

	case *command.CallbackTrigger:
		bot.OnCallbackQuery(
			action.Handler,
			buildCallbackPredicates(action, t, permissions)...,
		)
	}
}

func buildCommandPredicates(
	action *command.Action,
	trigger *command.CommandTrigger,
	permissions *predicate.PermissionChecker,
	rules *predicate.RuleChecker,
) []botapi.Predicate {
	var predicates []botapi.Predicate

	switch trigger.Scope {
	case command.ScopeGroup:
		predicates = append(predicates, predicate.Chat())

	case command.ScopePrivate:
		predicates = append(predicates, predicate.Private())

	case command.ScopeAny:
	}

	predicates = append(predicates,
		predicate.Command(action.Key, trigger.Prefixes, trigger.Aliases),
	)

	if len(trigger.Rules) != 0 {
		predicates = append(predicates,
			rules.With(trigger.Rules...),
		)
	} else {
		predicates = append(predicates, predicate.NoArgs())
	}

	if action.IgnorePermissionDenied {
		predicates = append(predicates,
			permissions.Pass(action.Key, action.Permission),
		)
	} else {
		predicates = append(predicates,
			permissions.Require(action.Key, action.Permission),
		)
	}

	predicates = append(predicates, action.ExtraPredicates...)

	return predicates
}

func buildCallbackPredicates(
	action *command.Action,
	trigger *command.CallbackTrigger,
	permissions *predicate.PermissionChecker,
) []botapi.Predicate {
	var predicates []botapi.Predicate

	if trigger.Prefix {
		predicates = append(predicates,
			botapi.CallbackPrefix(trigger.Data),
		)
	} else {
		predicates = append(predicates,
			botapi.CallbackData(trigger.Data),
		)
	}

	if action.IgnorePermissionDenied {
		predicates = append(predicates,
			permissions.Pass(action.Key, action.Permission),
		)
	} else {
		if action.AllowDev {
			predicates = append(predicates, permissions.PassDev())
		}
		predicates = append(predicates,
			permissions.Require(action.Key, action.Permission),
		)
	}

	predicates = append(predicates, action.ExtraPredicates...)

	return predicates
}
func allRulesAreOptional(rules []rule.Rule) bool {
	for _, r := range rules {
		if !r.IsOptional {
			return false
		}
	}

	return true
}

func BotCommands(registry *command.Registry, loc *i18n.Localizer, scope command.Scope, forAdmins bool) (botCommands []botapi.BotCommand) {
	for _, action := range registry.All() {
		if trigger, ok := action.Trigger.(*command.CommandTrigger); ok && (len(trigger.Rules) > 0 && !allRulesAreOptional(trigger.Rules)) {
			continue
		}

		if !forAdmins && action.Permission != permission.StatusMember {
			continue
		}

		cmd, ok := action.Trigger.(*command.CommandTrigger)
		if !ok || !cmd.Scope.BelongsTo(scope) {
			continue
		}

		botCommand := botapi.BotCommand{
			Command:     action.Key,
			Description: loc.T(action.Description, nil),
		}

		botCommands = append(botCommands, botCommand)
	}

	return
}

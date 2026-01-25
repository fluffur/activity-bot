package command

import (
	"strings"
	"unicode/utf8"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Command struct {
	Command  string
	Triggers []rune
	Aliases  []string
	Response Response
	MaxArgs  int
}

func NewCommand(c string, r Response, aliases ...string) Command {
	return Command{
		Command:  strings.ToLower(c),
		Triggers: []rune("/!."),
		Aliases:  aliases,
		Response: r,
	}
}

func (c Command) SetMaxArgs(maxArgs int) Command {
	c.MaxArgs = maxArgs
	return c
}

func (c Command) SetTriggers(triggers []rune) Command {
	c.Triggers = triggers
	return c
}

func (c Command) SetAliases(aliases ...string) Command {
	c.Aliases = aliases
	return c
}

func (c Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if ctx.Message != nil {
		if ctx.Message.GetText() == "" {
			return false
		}
		return c.checkMessage(b, ctx.Message)
	}

	return false
}
func (c Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	return c.Response(b, ctx, c.parseArgs(ctx.Message.GetText()))
}

func (c Command) parseArgs(text string) []string {
	var args []string
	lower := strings.ToLower(text)
	textRunes := []rune(lower)

	commands := append([]string{c.Command}, c.Aliases...)

	for _, t := range c.Triggers {
		for _, cmd := range commands {
			fullCmd := string(t) + strings.ToLower(cmd)

			if !strings.HasPrefix(lower, fullCmd) {
				continue
			}

			rest := strings.TrimSpace(string(textRunes[len(fullCmd):]))
			if rest == "" {
				return args
			}

			words := strings.Fields(rest)

			if c.MaxArgs <= 0 {
				return words
			}

			if len(words) <= c.MaxArgs {
				return []string{rest}
			}

			head := words[:len(words)-c.MaxArgs+1]
			tail := words[len(words)-c.MaxArgs+1:]

			args = append(args, strings.Join(append(head, tail...), " "))

			return args
		}
	}

	return args
}

func (c Command) Name() string {
	return "command_" + c.Command
}

func (c Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	ents := msg.GetEntities()
	if len(ents) != 0 && ents[0].Offset == 0 && ents[0].Type != "bot_command" {
		return false
	}

	text := msg.GetText()

	var cmd string
	for _, t := range c.Triggers {
		if r, _ := utf8.DecodeRuneInString(text); r != t {
			continue
		}

		split := strings.Split(strings.ToLower(strings.Fields(text)[0]), "@")
		if len(split) > 1 && split[1] != strings.ToLower(b.User.Username) {
			return false
		}
		cmd = split[0][1:]
		break
	}
	if cmd == "" {
		return false
	}

	if cmd == c.Command {
		return true
	}

	for _, alias := range c.Aliases {
		if cmd == strings.ToLower(alias) {
			return true
		}
	}

	return false
}

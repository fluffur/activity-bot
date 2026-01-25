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
	return c.Response(b, ctx, c.parseArgs(ctx.Message))
}

func (c Command) parseArgs(msg *gotgbot.Message) []string {
	text := msg.GetText()
	textRunes := []rune(text)
	lower := strings.ToLower(text)

	var rest string
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		rest = strings.TrimSpace(text)
	} else {
		restRunes := textRunes
		for _, e := range msg.Entities {
			if e.Type == "text_mention" || e.User != nil || e.Type == "mention" {
				start := int(e.Offset)
				end := start + int(e.Length)
				if start >= 0 && end <= len(restRunes) {
					restRunes = append(restRunes[:start], restRunes[end:]...)
				}
			}
		}
		rest = strings.TrimSpace(string(restRunes))
	}

	commands := append([]string{c.Command}, c.Aliases...)
	for _, t := range c.Triggers {
		for _, cmd := range commands {
			fullCmd := string(t) + strings.ToLower(cmd)
			if !strings.HasPrefix(lower, fullCmd) {
				continue
			}

			restRunes := []rune(rest)
			if strings.HasPrefix(strings.ToLower(rest), fullCmd) {
				restRunes = restRunes[len([]rune(fullCmd)):]
				rest = strings.TrimSpace(string(restRunes))
			}
			if rest == "" {
				return nil
			}

			if c.MaxArgs <= 0 {
				return strings.Fields(rest)
			}

			words := strings.Fields(rest)
			if len(words) <= c.MaxArgs {
				return []string{rest}
			}

			args := make([]string, 0, c.MaxArgs)
			for i := 0; i < c.MaxArgs-1; i++ {
				args = append(args, words[i])
			}
			last := strings.Join(words[c.MaxArgs-1:], " ")
			args = append(args, last)

			return args
		}
	}

	return nil
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

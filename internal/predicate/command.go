package predicate

import (
	"activity-bot/internal/middleware/cctx"
	"context"
	"strings"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

var defaultPrefixes = []string{"!", "/", ".", "фм"}

type argsKey struct{}

var commandArgsKey = argsKey{}

func Args(c *botapi.Context) string {
	if val, ok := c.Context.Value(commandArgsKey).(string); ok {
		return val
	}

	return ""
}

func Command(name string, aliases ...string) botapi.Predicate {
	commands := make([]string, 0, len(aliases)+1)

	commands = append(commands, normalize(name))

	for _, alias := range aliases {
		commands = append(commands, normalize(alias))
	}

	return func(c *botapi.Context) bool {
		m := c.Message()
		if m == nil {
			return false
		}

		ch, err := cctx.Chat(c.Context)
		if err != nil {
			log.For(c.Bot.Logger()).Error(
				c.Context,
				"command ctx chat",
				log.Error(err),
			)

			return false
		}

		text, _ := m.TextAndEntities()

		prefixes := append([]string{}, defaultPrefixes...)
		if ch.CommandPrefix != "" {
			prefixes = append(prefixes, ch.CommandPrefix)
		}

		prefix := findPrefix(text, prefixes)

		if !ch.AllowPrefixless && prefix == "" {
			return false
		}

		text = strings.TrimSpace(text[len(prefix):])

		if text == "" {
			return false
		}

		botUsername := strings.ToLower(c.Bot.Self().Username)

		for _, cmd := range commands {
			args, ok := parseCommand(text, cmd, botUsername)
			if !ok {
				continue
			}

			c.Context = context.WithValue(
				c.Context,
				commandArgsKey,
				args,
			)

			return true
		}

		return false
	}
}

func parseCommand(
	originalText string,
	command string,
	botUsername string,
) (string, bool) {
	words := strings.Fields(originalText)

	if len(words) == 0 {
		return "", false
	}

	cmdWords := strings.Fields(command)

	if len(words) < len(cmdWords) {
		return "", false
	}

	for i, cmdWord := range cmdWords {
		word := strings.ToLower(words[i])

		last := i == len(cmdWords)-1

		if last {
			if word == cmdWord {
				continue
			}

			if strings.HasPrefix(word, cmdWord+"@") {
				username := strings.TrimPrefix(word, cmdWord+"@")

				if username == botUsername {
					continue
				}
			}

			return "", false
		}

		if word != cmdWord {
			return "", false
		}
	}

	pos := len(cmdWords)

	if len(words) > pos && strings.HasPrefix(words[pos], "@") {
		if strings.ToLower(words[pos][1:]) != botUsername {
			return "", false
		}

		pos++
	}

	if len(words) <= pos {
		return "", true
	}

	return strings.Join(words[pos:], " "), true
}

func normalize(s string) string {
	return strings.Join(
		strings.Fields(strings.ToLower(s)),
		" ",
	)
}

func findPrefix(text string, prefixes []string) string {
	textLower := strings.ToLower(text)

	longest := ""

	for _, prefix := range prefixes {
		if strings.HasPrefix(textLower, strings.ToLower(prefix)) &&
			len(prefix) > len(longest) {
			longest = prefix
		}
	}

	return longest
}

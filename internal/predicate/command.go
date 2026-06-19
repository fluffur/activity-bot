package predicate

import (
	"activity-bot/internal/middleware/cctx"
	"context"
	"strings"

	"github.com/gotd/botapi"
)

var prefixes = []string{"!", "/", ".", "фм"}

type ArgsKey struct{}

func GetArgs(c *botapi.Context) string {
	if val, ok := c.Context.Value(ArgsKey{}).(string); ok {
		return val
	}
	return ""
}

func Command(name string, aliases ...string) botapi.Predicate {
	return func(c *botapi.Context) bool {
		m := c.Message()
		if m == nil {
			return false
		}

		ch, err := cctx.Chat(c.Context)
		if err != nil {
			return false
		}

		if ch.CommandPrefix != "" {
			prefixes = append(prefixes, ch.CommandPrefix)
		}

		originalText := m.Text
		if originalText == "" {
			originalText = m.Caption
		}

		textLower := strings.ToLower(originalText)

		prefixLower := findPrefix(textLower, prefixes)
		if !ch.AllowPrefixless && prefixLower == "" {
			return false
		}

		textNoPrefix := originalText[len(prefixLower):]
		textNoPrefixTrimmed := strings.TrimSpace(textNoPrefix)

		textLowerNormalized := strings.Join(strings.Fields(strings.ToLower(textNoPrefixTrimmed)), " ")

		myBotUsername := "@" + strings.ToLower(c.Bot.Self().Username)

		matchCommand := func(targetCmd string) (bool, string) {
			targetCmd = strings.ToLower(targetCmd)

			if !strings.HasPrefix(textLowerNormalized, targetCmd) {
				return false, ""
			}

			restLower := strings.TrimPrefix(textLowerNormalized, targetCmd)

			if restLower == "" {
				return true, ""
			}

			if strings.HasPrefix(restLower, " ") {
				return true, extractRawArgs(textNoPrefixTrimmed, targetCmd, "")
			}

			if strings.HasPrefix(restLower, "@") {
				parts := strings.SplitN(restLower, " ", 2)
				mentionedBot := parts[0]

				if mentionedBot != myBotUsername {
					return false, ""
				}

				if len(parts) == 1 {
					return true, ""
				}

				return true, extractRawArgs(textNoPrefixTrimmed, targetCmd, myBotUsername)
			}

			return false, ""
		}

		matched := false
		rawArgs := ""

		if matched, rawArgs = matchCommand(name); !matched {
			for _, alias := range aliases {
				if matched, rawArgs = matchCommand(alias); matched {
					break
				}
			}
		}

		if matched {
			c.Context = context.WithValue(c.Context, ArgsKey{}, rawArgs)
			return true
		}

		return false
	}
}

func extractRawArgs(originalNoPrefix, targetCmdLower, myBotUsername string) string {
	lower := strings.ToLower(originalNoPrefix)
	idx := strings.Index(lower, targetCmdLower)
	if idx == -1 {
		return strings.TrimSpace(originalNoPrefix)
	}

	startArgs := idx + len(targetCmdLower)

	if myBotUsername != "" && startArgs < len(lower) && lower[startArgs] == '@' {
		startArgs += len(myBotUsername)
	}

	return strings.TrimSpace(originalNoPrefix[startArgs:])
}

func findPrefix(textLower string, prefixes []string) string {
	for _, prefix := range prefixes {
		pLower := strings.ToLower(prefix)
		if strings.HasPrefix(textLower, pLower) {
			return prefix
		}
	}
	return ""
}

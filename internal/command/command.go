package command

import (
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Guard interface {
	Check(ctx *ext.Context) (bool, string)
}

type GuardFunc func(ctx *ext.Context) (bool, string)

func (f GuardFunc) Check(ctx *ext.Context) (bool, string) {
	return f(ctx)
}

type Command struct {
	command          string
	triggers         []string
	aliases          []string
	response         Response
	maxArgs          int
	allowArgs        bool
	fallbackToSender bool
	userService      *user.Service
	guards           []Guard
}

func New(c string, r Response, userService *user.Service, aliases ...string) *Command {
	return &Command{
		command:          strings.ToLower(c),
		triggers:         []string{"/", "!", "."},
		aliases:          aliases,
		response:         r,
		fallbackToSender: false,

		userService: userService,
		guards:      make([]Guard, 0),
	}
}

func (c *Command) WithGuards(guards ...Guard) *Command {
	c.guards = append(c.guards, guards...)
	return c
}

func (c *Command) FallbackToSender() *Command {
	c.fallbackToSender = true
	return c
}

func (c *Command) AllowArgs() *Command {
	c.allowArgs = true
	return c
}

func (c *Command) SetMaxArgs(maxArgs int) *Command {
	c.maxArgs = maxArgs
	return c
}

func (c *Command) SetTriggers(triggers ...string) *Command {
	c.triggers = triggers
	return c
}

func (c *Command) SetAliases(aliases ...string) *Command {
	c.aliases = aliases
	return c
}

func (c *Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if ctx.Message != nil {
		if ctx.Message.GetText() == "" {
			return false
		}
		return c.checkMessage(b, ctx.Message)
	}
	return false
}

func (c *Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	for _, guard := range c.guards {
		if ok, message := guard.Check(ctx); !ok {
			if message != "" {
				_, err := ctx.EffectiveMessage.Reply(b, message, nil)
				return err
			}
			return nil
		}
	}

	return c.response(b, ctx, c.parseArgs(b, ctx))
}

func (c *Command) ensureUser(u *gotgbot.User) (model.User, error) {
	return c.userService.EnsureUserExists(u.Id, u.Username, u.FirstName, u.LastName)

}
func (c *Command) parseArgs(b *gotgbot.Bot, ctx *ext.Context) *Context {
	msg := ctx.Message
	text := msg.GetText()
	textRunes := []rune(text)

	usersMap := make(map[int64]*model.User)
	removeRanges := make([][2]int, 0)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && !msg.ReplyToMessage.From.IsBot {
		u, err := c.ensureUser(msg.ReplyToMessage.From)
		if err == nil {
			usersMap[u.ID] = &u
		} else {
			log.Println("Ensure user from reply exists", err)
		}
	}

	for _, e := range msg.Entities {
		start := int(e.Offset)
		end := start + int(e.Length)
		if start < 0 || end > len(textRunes) {
			continue
		}

		switch e.Type {
		case "text_mention":
			if e.User != nil {
				u, err := c.ensureUser(e.User)
				if err == nil {
					usersMap[u.ID] = &u
				} else {
					log.Println("Ensure user from mention exists", err)
				}
				removeRanges = append(removeRanges, [2]int{start, end})
			}
		case "mention":
			username := string(textRunes[start+1 : end])
			u, err := c.userService.GetUserByUsername(username)
			if err == nil {
				usersMap[u.ID] = &u
			} else {
				log.Println("Ensure user from username mention exists", err)
			}
			removeRanges = append(removeRanges, [2]int{start, end})
		}
	}

	if c.fallbackToSender && len(usersMap) == 0 {
		u, err := c.userService.EnsureUserExists(ctx.EffectiveUser.Id, ctx.EffectiveUser.Username, ctx.EffectiveUser.FirstName, ctx.EffectiveUser.LastName)
		if err != nil {
			log.Println("Show EnsureUserExists failed", err)
		} else {
			usersMap[u.ID] = &u
		}
	}

	for i := len(removeRanges) - 1; i >= 0; i-- {
		r := removeRanges[i]
		textRunes = append(textRunes[:r[0]], textRunes[r[1]:]...)
	}

	rest := strings.TrimSpace(string(textRunes))

commandsLoop:
	for _, t := range c.triggers {
		for _, cmd := range append([]string{c.command}, c.aliases...) {
			fullCmd := string(t) + strings.ToLower(cmd)
			fullCmdWithBot := fullCmd + "@" + strings.ToLower(b.User.Username)

			if strings.HasPrefix(strings.ToLower(rest), fullCmd) || strings.HasPrefix(strings.ToLower(rest), fullCmdWithBot) {
				if strings.HasPrefix(strings.ToLower(rest), fullCmdWithBot) {
					rest = strings.TrimSpace(rest[len(fullCmdWithBot):])
				} else {
					rest = strings.TrimSpace(rest[len(fullCmd):])
				}
				break commandsLoop
			}
		}
	}

	words := strings.Fields(rest)
	if c.maxArgs > 0 && len(words) > c.maxArgs {
		last := strings.Join(words[c.maxArgs-1:], " ")
		words = append(words[:c.maxArgs-1], last)
	}

	users := make([]*model.User, 0, len(usersMap))
	for _, u := range usersMap {
		users = append(users, u)
	}

	return &Context{
		Args:  words,
		Users: users,
	}
}

func (c *Command) Name() string {
	return "command_" + c.command
}

func (c *Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	text := strings.ToLower(removeMentions(msg))
	if text == "" {
		return false
	}

	for _, trigger := range c.triggers {
		if !strings.HasPrefix(text, trigger) {
			continue
		}

		for _, cName := range append([]string{c.command}, c.aliases...) {
			fullCmd := trigger + strings.ToLower(cName)
			fullCmdWithBot := fullCmd + "@" + strings.ToLower(b.User.Username)

			if strings.HasPrefix(text, fullCmd) || strings.HasPrefix(text, fullCmdWithBot) {
				rest := text[len(fullCmd):]

				if strings.HasPrefix(rest, "@"+strings.ToLower(b.User.Username)) {
					rest = rest[len(b.User.Username)+1:]
				}

				rest = strings.TrimSpace(rest)

				if !c.allowArgs && len(rest) > 0 {
					return false
				}
				return true
			}
		}
	}

	return false
}

func removeMentions(msg *gotgbot.Message) string {
	text := msg.GetText()
	textRunes := []rune(text)

	removeRanges := make([][2]int, 0)

	for _, e := range msg.Entities {
		start := int(e.Offset)
		end := start + int(e.Length)
		if start < 0 || end > len(textRunes) {
			continue
		}

		switch e.Type {
		case "mention", "text_mention":
			removeRanges = append(removeRanges, [2]int{start, end})
		}
	}

	for i := len(removeRanges) - 1; i >= 0; i-- {
		r := removeRanges[i]
		textRunes = append(textRunes[:r[0]], textRunes[r[1]:]...)
	}

	return strings.TrimSpace(string(textRunes))
}

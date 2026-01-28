package cmd

import (
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const ArgsCountAny = -1

type Command struct {
	command          string
	triggers         []string
	aliases          []string
	response         Response
	argsCount        int
	fallbackToSender bool
	userService      *user.Service
	guards           []Guard
}

func New(c string, r Response, userService *user.Service) *Command {
	return &Command{
		command:          strings.ToLower(c),
		triggers:         []string{"/", "!", "."},
		aliases:          []string{},
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

func (c *Command) SetArgsCount(argsCount int) *Command {
	c.argsCount = argsCount
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
	usersMap := make(map[int64]*model.User)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && !msg.ReplyToMessage.From.IsBot {
		u, err := c.ensureUser(msg.ReplyToMessage.From)
		if err == nil {
			usersMap[u.ID] = &u
		} else {
			log.Println("Ensure user from reply exists", err)
		}
	}

	text, entities := cleanMessage(msg)
	textRunes := []rune(msg.GetText())

	for _, e := range entities {
		switch e.Type {
		case "text_mention":
			if e.User != nil {
				u, err := c.ensureUser(e.User)
				if err == nil {
					usersMap[u.ID] = &u
				} else {
					log.Println("Ensure user from mention exists", err)
				}
			}
		case "mention":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(textRunes) {
				username := string(textRunes[start+1 : end])
				u, err := c.userService.GetUserByUsername(username)
				if err == nil {
					usersMap[u.ID] = &u
				} else {
					log.Println("Ensure user from username mention exists", err)
				}
			}
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

	rest, matched := c.matchCommand(text, b.User.Username)
	if !matched {
		log.Println("Command logic mismatch: matchCommand failed in parseArgs")
		return &Context{args: []string{}, users: []*model.User{}}
	}
	words := strings.Fields(rest)
	if c.argsCount != ArgsCountAny && c.argsCount > 0 && len(words) > c.argsCount {
		last := strings.Join(words[c.argsCount-1:], " ")
		words = append(words[:c.argsCount-1], last)
	}

	users := make([]*model.User, 0, len(usersMap))
	for _, u := range usersMap {
		users = append(users, u)
	}

	return &Context{
		args:  words,
		users: users,
	}
}

func (c *Command) Name() string {
	return "command_" + c.command
}

func (c *Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	text, _ := cleanMessage(msg)
	if text == "" {
		return false
	}

	rest, matched := c.matchCommand(text, b.User.Username)
	if !matched {
		return false
	}

	if c.argsCount == 0 && len(rest) > 0 {
		return false
	}

	return true
}

func (c *Command) matchCommand(text string, botUsername string) (string, bool) {
	text = strings.ToLower(text)
	botUsername = strings.ToLower(botUsername)

	for _, t := range c.triggers {
		for _, cmd := range append([]string{c.command}, c.aliases...) {
			fullCmd := t + strings.ToLower(cmd)
			fullCmdWithBot := fullCmd + "@" + botUsername

			if strings.HasPrefix(text, fullCmdWithBot) {
				return strings.TrimSpace(text[len(fullCmdWithBot):]), true
			}
			if strings.HasPrefix(text, fullCmd) {
				return strings.TrimSpace(text[len(fullCmd):]), true
			}
		}
	}
	return "", false
}

func cleanMessage(msg *gotgbot.Message) (string, []gotgbot.MessageEntity) {
	text := msg.GetText()
	textRunes := []rune(text)
	removeRanges := make([][2]int, 0)
	removedEntities := make([]gotgbot.MessageEntity, 0)

	for _, e := range msg.Entities {
		start := int(e.Offset)
		end := start + int(e.Length)
		if start < 0 || end > len(textRunes) {
			continue
		}

		switch e.Type {
		case "mention", "text_mention":
			removeRanges = append(removeRanges, [2]int{start, end})
			removedEntities = append(removedEntities, e)
		}
	}

	for i := len(removeRanges) - 1; i >= 0; i-- {
		r := removeRanges[i]
		textRunes = append(textRunes[:r[0]], textRunes[r[1]:]...)
	}

	return strings.TrimSpace(string(textRunes)), removedEntities
}

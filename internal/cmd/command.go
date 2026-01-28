package cmd

import (
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"log"
	"strings"
	"unicode/utf16"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const ArgsCountAny = -1
const ArgsCountNone = 0

var defaultTriggers = []string{"/"}

type Factory struct {
	userService *user.Service
	triggers    []string
}

func NewFactory(userService *user.Service, triggers ...string) *Factory {
	if len(triggers) == 0 {
		triggers = defaultTriggers
	}

	return &Factory{userService, triggers}
}

func (f *Factory) New(r Response, c string, aliases ...string) *Command {
	return New(append(aliases, c), f.triggers, r, f.userService)
}

type Command struct {
	commands         []string
	triggers         []string
	response         Response
	argsCount        int
	fallbackToSender bool
	userService      *user.Service
	guards           []Guard
}

func New(commands []string, triggers []string, response Response, userService *user.Service) *Command {
	lowerCommands := make([]string, len(commands))
	for i, c := range commands {
		lowerCommands[i] = strings.ToLower(c)
	}

	return &Command{
		commands:         lowerCommands,
		triggers:         triggers,
		response:         response,
		fallbackToSender: false,
		argsCount:        ArgsCountNone,
		userService:      userService,
		guards:           make([]Guard, 0),
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

func (c *Command) AddAliases(aliases ...string) *Command {
	for _, a := range aliases {
		c.commands = append(c.commands, strings.ToLower(a))
	}
	return c
}

func (c *Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if ctx.Message == nil || ctx.Message.GetText() == "" {
		return false
	}
	return c.checkMessage(b, ctx.Message)
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
	users := make([]*model.User, 0)
	seen := make(map[int64]bool)

	addUser := func(u *model.User) {
		if u != nil && !seen[u.ID] {
			users = append(users, u)
			seen[u.ID] = true
		}
	}

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && !msg.ReplyToMessage.From.IsBot {
		u, err := c.ensureUser(msg.ReplyToMessage.From)
		if err == nil {
			addUser(&u)
		} else {
			log.Println("Ensure user from reply exists failed:", err)
		}
	}

	text, entities := extractMentions(msg)
	utf16Text := utf16.Encode([]rune(msg.GetText()))

	for _, e := range entities {
		switch e.Type {
		case "text_mention":
			if e.User != nil {
				u, err := c.ensureUser(e.User)
				if err == nil {
					addUser(&u)
				} else {
					log.Println("Ensure user from text_mention exists failed:", err)
				}
			}
		case "mention":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(utf16Text) {
				username := string(utf16.Decode(utf16Text[start+1 : end]))
				u, err := c.userService.GetUserByUsername(username)
				if err == nil {
					addUser(&u)
				} else {
					log.Println("Get user by username from mention failed:", err)
				}
			}
		}
	}

	if c.fallbackToSender && len(users) == 0 {
		u, err := c.ensureUser(ctx.EffectiveUser)
		if err == nil {
			addUser(&u)
		} else {
			log.Println("Ensure sender exists failed:", err)
		}
	}

	rest, _ := c.matchCommand(text, b.User.Username)
	words := strings.Fields(rest)

	if c.argsCount != ArgsCountAny && c.argsCount > 0 && len(words) > c.argsCount {
		lastArgStart := 0
		tempRest := rest
		for i := 0; i < c.argsCount-1; i++ {
			word := words[i]
			idx := strings.Index(tempRest, word)
			if idx == -1 {
				break
			}
			offset := idx + len(word)
			lastArgStart += offset
			tempRest = tempRest[offset:]
		}

		last := strings.TrimSpace(rest[lastArgStart:])
		words = append(words[:c.argsCount-1], last)
	}

	return &Context{
		args:  words,
		users: users,
	}
}

func (c *Command) Name() string {
	if len(c.commands) > 0 {
		return "command_" + c.commands[0]
	}
	return "unnamed_command"
}

func (c *Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	text, _ := extractMentions(msg)
	if text == "" {
		return false
	}

	rest, matched := c.matchCommand(text, b.User.Username)
	if !matched {
		return false
	}

	if c.argsCount == ArgsCountNone && len(strings.TrimSpace(rest)) > 0 {
		return false
	}

	return true
}

func (c *Command) matchCommand(text string, botUsername string) (string, bool) {
	textLower := strings.ToLower(text)
	botUsername = strings.ToLower(botUsername)

	for _, t := range c.triggers {
		for _, cmd := range c.commands {
			prefixWithBot := strings.ToLower(t + cmd + "@" + botUsername)
			if strings.HasPrefix(textLower, prefixWithBot) {
				return c.extractRest(text, len(prefixWithBot))
			}

			prefix := strings.ToLower(t + cmd)
			if strings.HasPrefix(textLower, prefix) {
				return c.extractRest(text, len(prefix))
			}
		}
	}
	return "", false
}

func (c *Command) extractRest(text string, prefixLen int) (string, bool) {
	rest := text[prefixLen:]
	if len(rest) == 0 {
		return "", true
	}

	first := rest[0]
	if first == ' ' || first == '\n' || first == ',' || first == '\t' {
		return strings.TrimSpace(rest), true
	}

	return "", false
}

func extractMentions(msg *gotgbot.Message) (string, []gotgbot.MessageEntity) {
	text := msg.GetText()
	utf16Text := utf16.Encode([]rune(text))
	removeRanges := make([][2]int, 0)
	removedEntities := make([]gotgbot.MessageEntity, 0)

	for _, e := range msg.Entities {
		if e.Type == "mention" || e.Type == "text_mention" {
			removeRanges = append(removeRanges, [2]int{int(e.Offset), int(e.Offset + e.Length)})
			removedEntities = append(removedEntities, e)
		}
	}

	for i := len(removeRanges) - 1; i >= 0; i-- {
		r := removeRanges[i]
		if r[0] >= 0 && r[1] <= len(utf16Text) {
			utf16Text = append(utf16Text[:r[0]], utf16Text[r[1]:]...)
		}
	}

	return strings.TrimSpace(string(utf16.Decode(utf16Text))), removedEntities
}

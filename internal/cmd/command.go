package cmd

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"log"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const ArgsCountAny = -1
const ArgsCountNone = 0

var defaultTriggers = []string{"/"}

type Factory struct {
	userService    *user.Service
	chatService    *chat.Service
	sessionService interface {
		GetActiveChat(ctx context.Context, userID int64) (int64, error)
	}
	uniquePrefix string
	triggers     []string
}

func NewFactory(userService *user.Service, chatService *chat.Service, sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}, uniquePrefix string, triggers ...string) *Factory {
	if len(triggers) == 0 {
		triggers = defaultTriggers
	}

	for i, t := range triggers {
		triggers[i] = strings.ToLower(t)
	}

	return &Factory{userService, chatService, sessionService, strings.ToLower(uniquePrefix), triggers}
}

func (f *Factory) New(r Response, c string, aliases ...string) *Command {
	return New(append(aliases, c), f.triggers, r, f.userService, f.chatService, f.sessionService, f.uniquePrefix)
}

type Command struct {
	commands         []string
	triggers         []string
	response         Response
	argsCount        int
	fallbackToSender bool
	userService      *user.Service
	chatService      *chat.Service
	sessionService   interface {
		GetActiveChat(ctx context.Context, userID int64) (int64, error)
	}
	uniquePrefix string
	guards       []Guard
	forcePrefix  bool
}

func New(commands []string, triggers []string, response Response, userService *user.Service, chatService *chat.Service, sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}, uniquePrefix string) *Command {
	for i, c := range commands {
		commands[i] = strings.ToLower(c)
	}

	return &Command{
		commands:         commands,
		triggers:         triggers,
		response:         response,
		fallbackToSender: false,
		argsCount:        ArgsCountNone,
		userService:      userService,
		chatService:      chatService,
		sessionService:   sessionService,
		uniquePrefix:     uniquePrefix,
		guards:           make([]Guard, 0),
		forcePrefix:      false,
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

func (c *Command) ForcePrefix() *Command {
	c.forcePrefix = true
	return c
}

func (c *Command) SetArgsCount(argsCount int) *Command {
	if argsCount < 0 && argsCount != ArgsCountAny {
		return c
	}

	c.argsCount = argsCount
	return c
}

func (c *Command) SetTriggers(triggers ...string) *Command {
	c.triggers = triggers
	return c
}

func (c *Command) AddTriggers(trigger ...string) *Command {
	c.triggers = append(c.triggers, trigger...)
	return c
}

func (c *Command) AddAliases(aliases ...string) *Command {
	for _, a := range aliases {
		c.commands = append(c.commands, strings.ToLower(a))
	}
	return c
}

func (c *Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if ctx.Message == nil || ctx.Message.ForwardOrigin != nil || ctx.Message.GetText() == "" {
		return false
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return c.checkMessage(ctxWithTimeout, b, ctx.Message)
}

func (c *Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, guard := range c.guards {
		if ok, message := guard.Check(ctx, c.commands[0], ctxWithTimeout); !ok {
			if message != "" {
				_, err := ctx.EffectiveMessage.Reply(b, message, nil)
				return err
			}
			return nil
		}
	}

	return c.response(b, c.parseArgs(b, ctx, ctxWithTimeout))
}

func (c *Command) ensureUser(ctx context.Context, u *gotgbot.User) (model.User, error) {
	return c.userService.EnsureUserExists(ctx, u.Id, u.Username, u.FirstName, u.LastName)
}
func (c *Command) parseArgs(b *gotgbot.Bot, ctx *ext.Context, cctx context.Context) *Context {
	msg := ctx.Message
	users := make([]*model.User, 0)
	takenUsers := make(map[int64]struct{})

	var fullHTML string
	if msg.Caption != "" {
		fullHTML = msg.OriginalCaptionHTML()
	} else {
		fullHTML = msg.OriginalHTML()
	}

	addUser := func(user *model.User) {
		if _, ok := takenUsers[user.ID]; ok {
			return
		}
		users = append(users, user)
		takenUsers[user.ID] = struct{}{}
	}

	text, entities := cleanMessage(msg)
	textRunes := []rune(msg.GetText())
	for _, e := range entities {
		switch e.Type {
		case "text_mention":
			if e.User != nil {
				u, err := c.ensureUser(cctx, e.User)
				if err == nil {
					addUser(&u)
				}
			}
		case "mention":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(textRunes) {
				username := string(textRunes[start+1 : end])
				u, err := c.userService.GetUserByUsername(cctx, username)
				if err == nil {
					addUser(&u)
				}
			}
		case "url":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(textRunes) {
				username := strings.TrimPrefix(strings.TrimPrefix(string(textRunes[start:end]), "https://"), "t.me/")
				u, err := c.userService.GetUserByUsername(cctx, username)
				if err == nil {
					addUser(&u)
				}
			}
		}
	}
	if msg.ReplyToMessage != nil &&
		msg.ReplyToMessage.ForumTopicCreated == nil &&
		msg.ReplyToMessage.From != nil &&
		!msg.ReplyToMessage.From.IsBot &&
		!msg.ReplyToMessage.IsAutomaticForward {
		u, err := c.ensureUser(cctx, msg.ReplyToMessage.From)
		if err == nil {
			addUser(&u)
		} else {
			log.Println(err)
		}
	}
	if c.fallbackToSender && len(users) == 0 {
		u, err := c.userService.EnsureUserExists(cctx, ctx.EffectiveUser.Id, ctx.EffectiveUser.Username, ctx.EffectiveUser.FirstName, ctx.EffectiveUser.LastName)
		if err == nil {
			addUser(&u)
		}
	}

	allowPrefixless := true
	chatPrefix := ""
	if c.chatService != nil {
		cht, err := c.chatService.GetChat(cctx, msg.Chat.Id)
		if err == nil {
			chatPrefix = strings.ToLower(cht.CommandPrefix)
			allowPrefixless = cht.AllowPrefixless
		}
	}

	rest, matched := c.matchCommand(text, b.User.Username, chatPrefix, allowPrefixless && !c.forcePrefix)
	if !matched {
		return &Context{ctx, cctx, []string{}, "", users, c.sessionService}
	}

	rest = strings.TrimSpace(rest)
	var htmlRest string

	if fullHTML != "" {
		plainText := msg.GetText()

		idx := strings.Index(plainText, rest)
		if idx >= 0 && idx <= len(fullHTML) {
			htmlRest = strings.TrimSpace(fullHTML[idx:])
		}
	}
	var args []string

	if c.argsCount == 2 {
		parts := strings.SplitN(rest, "\n", 2)
		args = append(args, strings.TrimSpace(parts[0]))
		if len(parts) > 1 {
			args = append(args, strings.TrimSpace(parts[1]))
		} else {
			args = append(args, "")
		}
	} else if c.argsCount == 1 || c.argsCount == ArgsCountNone || c.argsCount == ArgsCountAny {
		args = append(args, rest)
	}
	for _, u := range users {
		log.Println("user", *u)
	}
	return &Context{ctx, cctx, args, htmlRest, users, c.sessionService}
}

func (c *Command) Name() string {
	if len(c.commands) > 0 {
		return "command_" + c.commands[0]
	}
	return "unnamed_command"

}

func (c *Command) checkMessage(ctx context.Context, b *gotgbot.Bot, msg *gotgbot.Message) bool {
	text, _ := cleanMessage(msg)
	if text == "" {
		return false
	}

	allowPrefixless := true
	chatPrefix := ""
	if c.chatService != nil {
		cht, err := c.chatService.GetChat(ctx, msg.Chat.Id)
		if err == nil {
			chatPrefix = strings.ToLower(cht.CommandPrefix)
			allowPrefixless = cht.AllowPrefixless
		}
	}

	rest, matched := c.matchCommand(text, b.User.Username, chatPrefix, allowPrefixless && !c.forcePrefix)
	if !matched {
		return false
	}

	if c.argsCount == ArgsCountNone && len(rest) > 0 {
		return false
	}

	return true
}

func (c *Command) findTrigger(text string) (string, bool) {
	for _, t := range c.triggers {
		if strings.HasPrefix(text, t) {
			return t, true
		}
	}
	return "", false
}

func (c *Command) matchCommand(text string, botUsername string, chatPrefix string, allowPrefixless bool) (string, bool) {
	botUsername = strings.ToLower(botUsername)
	textLower := strings.ToLower(text)

	prefixes := make([]string, 0)
	if c.uniquePrefix != "" {
		prefixes = append(prefixes, c.uniquePrefix)
	}
	if chatPrefix != "" {
		prefixes = append(prefixes, chatPrefix)
	}

	for _, p := range prefixes {
		for _, t := range c.triggers {
			if strings.HasPrefix(textLower, t+p) {
				inner := strings.TrimSpace(text[len(t+p):])
				if rest, matched := c.matchCommandName(inner, botUsername); matched {
					return rest, true
				}
			}
		}
		if strings.HasPrefix(textLower, p) {
			inner := strings.TrimSpace(text[len(p):])
			if rest, matched := c.matchCommandName(inner, botUsername); matched {
				return rest, true
			}
		}
	}

	trigger, found := c.findTrigger(textLower)
	if !found {
		if !allowPrefixless {
			return "", false
		}
		return c.matchCommandName(text, botUsername)
	}

	if trigger == "" && !allowPrefixless {
		return "", false
	}

	textAfterTrigger := strings.TrimSpace(text[len(trigger):])
	return c.matchCommandName(textAfterTrigger, botUsername)
}

func (c *Command) matchCommandName(text string, botUsername string) (string, bool) {
	textLower := strings.ToLower(text)
	for _, cmd := range c.commands {
		if hasPrefix, sep := hasCommandPrefix(textLower, cmd); hasPrefix {
			rest := strings.TrimSpace(text[len(cmd+sep):])
			return rest, true
		}

		full := cmd + "@" + botUsername
		if hasPrefix, sep := hasCommandPrefix(textLower, full); hasPrefix {
			rest := strings.TrimSpace(text[len(full+sep):])
			return rest, true
		}
	}
	return "", false
}

func hasCommandPrefix(text, cmd string) (bool, string) {
	text = strings.ToLower(text)
	cmd = strings.ToLower(cmd)
	if text == cmd {
		return true, ""
	}
	separators := []string{" ", ",", "\n"}
	for _, sep := range separators {
		if strings.HasPrefix(text, cmd+sep) {
			return true, sep
		}
	}
	return false, ""
}

func cleanMessage(msg *gotgbot.Message) (string, []gotgbot.MessageEntity) {
	text := msg.GetText()
	removeRanges := make([][2]int, 0)
	removedEntities := make([]gotgbot.MessageEntity, 0)

	for _, e := range msg.Entities {
		start, end := runeIndex(text, int(e.Offset), int(e.Length))
		switch e.Type {
		case "mention", "text_mention", "url":
			removeRanges = append(removeRanges, [2]int{start, end})
			removedEntities = append(removedEntities, e)
		}
	}

	textRunes := []rune(text)
	for i := len(removeRanges) - 1; i >= 0; i-- {
		r := removeRanges[i]
		if r[0] >= 0 && r[1] <= len(textRunes) {
			textRunes = append(textRunes[:r[0]], textRunes[r[1]:]...)
		}
	}

	return strings.TrimSpace(string(textRunes)), removedEntities
}

func runeIndex(text string, offset, length int) (start, end int) {
	runes := []rune(text)
	utf16Pos := 0
	start = -1
	for i, r := range runes {
		if utf16Pos == offset {
			start = i
		}
		utf16Pos += utf16Length(r)
		if utf16Pos >= offset+length {
			end = i + 1
			break
		}
	}
	if start == -1 {
		start = 0
	}
	if end == 0 {
		end = len(runes)
	}
	return
}

func utf16Length(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

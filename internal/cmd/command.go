package cmd

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"strconv"
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
	memberService  *member.Service
	sessionService interface {
		GetActiveChat(ctx context.Context, userID int64) (int64, error)
	}
	uniquePrefix string
	triggers     []string
	commands     []*Command
}

func NewFactory(userService *user.Service, chatService *chat.Service, memberService *member.Service, sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}, uniquePrefix string, triggers ...string) *Factory {
	if len(triggers) == 0 {
		triggers = defaultTriggers
	}

	for i, t := range triggers {
		triggers[i] = strings.ToLower(t)
	}

	return &Factory{userService, chatService, memberService, sessionService, strings.ToLower(uniquePrefix), triggers, make([]*Command, 0)}
}

func (f *Factory) New(r Response, c string, aliases ...string) *Command {
	cmd := New(append(aliases, c), f.triggers, r, f.userService, f.chatService, f.memberService, f.sessionService, f.uniquePrefix)
	f.commands = append(f.commands, cmd)
	return cmd
}

func (f *Factory) WrapCallback(r Response, guards ...Guard) func(b *gotgbot.Bot, ctx *ext.Context) error {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		ctxWithTimeout := context.Background()

		for _, guard := range guards {
			if ok, message := guard.Check(ctx, "", ctxWithTimeout); !ok {
				if message != "" && ctx.CallbackQuery != nil {
					_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: message, ShowAlert: true})
				}
				return nil
			}
		}
		chatID, err := GetChatID(f.sessionService, ctx, ctxWithTimeout)
		if err != nil {
			logger.L.Warn("Chat ID not found for callback, use effective chat", "error", err)
			chatID = ctx.EffectiveChat.Id
		}
		cmdCtx := &Context{
			Context:      ctx,
			ctx:          ctxWithTimeout,
			targetChatID: chatID,
		}

		if ctx.CallbackQuery != nil {
			parts := strings.Split(ctx.CallbackQuery.Data, ":")
			if len(parts) > 1 {
				if userID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					m, err := f.memberService.GetChatMember(ctxWithTimeout, chatID, userID)
					if err == nil {
						cmdCtx.members = []*model.ChatMember{&m}

						if len(parts) > 2 {
							joinedArgs := strings.Join(parts[2:], ":")
							cmdCtx.args = strings.Fields(joinedArgs)
						} else {
							cmdCtx.args = []string{}
						}
					} else {
						u, err := f.userService.GetUser(ctxWithTimeout, userID)
						if err == nil {
							cmdCtx.members = []*model.ChatMember{{User: u}}
							if len(parts) > 2 {
								joinedArgs := strings.Join(parts[2:], ":")
								cmdCtx.args = strings.Fields(joinedArgs)
							} else {
								cmdCtx.args = []string{}
							}
						} else {
							cmdCtx.args = parts[1:]
						}
					}
				} else {
					cmdCtx.args = parts[1:]
				}
			}
		}

		err = r(b, cmdCtx)
		if err != nil {
			return err
		}
		if ctx.CallbackQuery != nil {
			_, err := ctx.CallbackQuery.Answer(b, nil)
			return err
		}
		return nil
	}
}

func (f *Factory) ConfigurableCommands() []*Command {
	var res []*Command
	for _, c := range f.commands {
		if c.description != "" {
			res = append(res, c)
		}
	}
	return res
}

func (f *Factory) WrapEvent(r Response, guards ...Guard) func(b *gotgbot.Bot, ctx *ext.Context) error {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		ctxWithTimeout := context.Background()

		for _, guard := range guards {
			if ok, message := guard.Check(ctx, "", ctxWithTimeout); !ok {
				if message != "" && ctx.EffectiveMessage != nil {
					_, _ = ctx.EffectiveMessage.Reply(b, message, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML})
				}
				return nil
			}
		}

		members := make([]*model.ChatMember, 0)
		if ctx.Message != nil {
			if ctx.Message.NewChatMembers != nil {
				for _, u := range ctx.Message.NewChatMembers {
					m, err := f.memberService.GetChatMember(ctxWithTimeout, ctx.EffectiveChat.Id, u.Id)
					if err == nil {
						members = append(members, &m)
					} else {
						mu, err := f.userService.EnsureUserExists(ctxWithTimeout, u.Id, u.Username, u.FirstName, u.LastName)
						if err == nil {
							members = append(members, &model.ChatMember{User: mu})
						}
					}
				}
			} else if ctx.Message.LeftChatMember != nil {
				u := ctx.Message.LeftChatMember
				m, err := f.memberService.GetChatMember(ctxWithTimeout, ctx.EffectiveChat.Id, u.Id)
				if err == nil {
					members = append(members, &m)
				} else {
					mu, err := f.userService.EnsureUserExists(ctxWithTimeout, u.Id, u.Username, u.FirstName, u.LastName)
					if err == nil {
						members = append(members, &model.ChatMember{User: mu})
					}
				}
			}
		}

		cmdCtx := &Context{
			Context:      ctx,
			ctx:          ctxWithTimeout,
			members:      members,
			targetChatID: ctx.EffectiveChat.Id,
		}

		return r(b, cmdCtx)
	}
}

type SessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}

type Command struct {
	commands         []string
	triggers         []string
	response         Response
	argsCount        int
	fallbackToSender bool
	userService      *user.Service
	chatService      *chat.Service
	memberService    *member.Service
	sessionService   SessionService
	uniquePrefix     string
	guards           []Guard
	forcePrefix      bool
	ambiguityPrefix  string
	requiredStatus   model.Status
	description      string
	category         Category
}

func New(commands []string, triggers []string, response Response, userService *user.Service, chatService *chat.Service, memberService *member.Service, sessionService interface {
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
		memberService:    memberService,
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

func (c *Command) WithAmbiguityResolution(prefix string) *Command {
	c.ambiguityPrefix = prefix
	return c
}

func (c *Command) Restricted(status model.Status) *Command {
	c.requiredStatus = status
	return c
}

func (c *Command) SetArgsCount(argsCount int) *Command {
	if argsCount < 0 && argsCount != ArgsCountAny {
		return c
	}

	c.argsCount = argsCount
	return c
}

func (c *Command) WithDescription(desc string) *Command {
	c.description = desc
	return c
}

func (c *Command) WithCategory(cat Category) *Command {
	c.category = cat
	return c
}

func (c *Command) GetKey() string {
	if len(c.commands) > 0 {
		return strings.ReplaceAll(c.commands[0], " ", "_")
	}
	return ""
}

func (c *Command) GetAliases() []string {
	return c.commands
}

func (c *Command) GetDescription() string {
	return c.description
}

func (c *Command) GetCategory() Category {
	return c.category
}

func (c *Command) GetDefaultStatus() model.Status {
	return c.requiredStatus
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
	ctxWithTimeout := context.Background()

	chatID, err := GetChatID(c.sessionService, ctx, ctxWithTimeout)
	if err != nil {
		logger.L.Error("Failed to get chat ID for permission check", "error", err)
		return nil
	}

	required := c.requiredStatus
	if c.chatService != nil {
		if s, err := c.chatService.GetCommandPermission(ctxWithTimeout, chatID, c.commands[0]); err == nil {
			required = s
		}
	}

	m, err := c.memberService.GetChatMember(ctxWithTimeout, chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return nil
	}

	if m.Status < required {
		message := fmt.Sprintf("%s Требуются права: %s", helpers.StatusEmoji(required), helpers.StatusName(required))
		_, err := ctx.EffectiveMessage.Reply(b, message, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML})
		return err
	}

	for _, guard := range c.guards {
		if ok, message := guard.Check(ctx, c.commands[0], ctxWithTimeout); !ok {
			if message != "" {
				_, err := ctx.EffectiveMessage.Reply(b, message, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML})
				return err
			}
			return nil
		}
	}

	cmdCtx := c.parseArgs(b, ctx, ctxWithTimeout)
	if c.ambiguityPrefix != "" && len(cmdCtx.members) > 1 {
		resolved, err := cmdCtx.ResolveUserAmbiguity(b, c.ambiguityPrefix, cmdCtx.ArgsString())
		if err != nil {
			return err
		}
		if resolved {
			return nil
		}
	}

	return c.response(b, cmdCtx)
}

func (c *Command) ensureMember(ctx context.Context, chatID int64, u *gotgbot.User) (model.ChatMember, error) {
	_, err := c.userService.EnsureUserExists(ctx, u.Id, u.Username, u.FirstName, u.LastName)
	if err != nil {
		return model.ChatMember{}, err
	}
	return c.memberService.GetChatMember(ctx, chatID, u.Id)
}

func (c *Command) parseArgs(b *gotgbot.Bot, ctx *ext.Context, cctx context.Context) *Context {
	msg := ctx.Message
	members := make([]*model.ChatMember, 0)
	parsedDates := make([]time.Time, 0)
	takenUsers := make(map[int64]struct{})

	var fullHTML string
	if msg.Caption != "" {
		fullHTML = msg.OriginalCaptionHTML()
	} else {
		fullHTML = msg.OriginalHTML()
	}

	addMember := func(m *model.ChatMember) {
		if _, ok := takenUsers[m.User.ID]; ok {
			return
		}
		members = append(members, m)
		takenUsers[m.User.ID] = struct{}{}
	}

	text, entities := cleanMessage(msg)
	textRunes := []rune(msg.GetText())
	chatID, err := c.getChatID(cctx, msg)
	if err != nil {
		logger.L.Error("getChatID error", "error", err)
		return &Context{ctx, cctx, []string{}, "", members, 0, parsedDates}
	}

	for _, e := range entities {
		switch e.Type {
		case "text_mention":
			if e.User != nil {
				m, err := c.ensureMember(cctx, chatID, e.User)
				if err == nil {
					addMember(&m)
				}
			}
		case "mention":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(textRunes) {
				username := string(textRunes[start+1 : end])
				m, err := c.memberService.GetChatMemberByUsername(cctx, chatID, username)
				if err == nil {
					addMember(&m)
				}
			}
		case "url":
			start := int(e.Offset)
			end := start + int(e.Length)
			if start >= 0 && end <= len(textRunes) {
				username := strings.TrimPrefix(strings.TrimPrefix(string(textRunes[start:end]), "https://"), "t.me/")
				m, err := c.memberService.GetChatMemberByUsername(cctx, chatID, username)
				if err == nil {
					addMember(&m)
				}
			}
		}
	}

	for _, e := range msg.GetEntities() {
		if e.Type != "date_time" {
			continue
		}

		if e.UnixTime != 0 {
			parsedDates = append(parsedDates, time.Unix(e.UnixTime, 0))
		}

	}

	if msg.ReplyToMessage != nil &&
		msg.ReplyToMessage.ForumTopicCreated == nil &&
		msg.ReplyToMessage.From != nil &&
		!msg.ReplyToMessage.From.IsBot &&
		!msg.ReplyToMessage.IsAutomaticForward {
		m, err := c.ensureMember(cctx, chatID, msg.ReplyToMessage.From)
		if err == nil {
			addMember(&m)
		} else {
			logger.L.Error("Failed to ensure reply member", "error", err)
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
		return &Context{ctx, cctx, []string{}, "", members, chatID, parsedDates}
	}

	rest = strings.TrimSpace(rest)

	if len(members) == 0 && rest != "" {
		strippedRest, foundMembers := c.findMembersByTitle(cctx, msg, rest)
		if len(foundMembers) > 0 {
			for _, m := range foundMembers {
				addMember(m)
			}
			rest = strippedRest
		}
	}

	if c.fallbackToSender && len(members) == 0 {
		m, err := c.ensureMember(cctx, chatID, ctx.EffectiveUser)
		if err == nil {
			addMember(&m)
		}
	}

	if fullHTML != "" {
		for _, t := range c.triggers {
			fullHTML = strings.TrimSpace(strings.TrimPrefix(fullHTML, t))
		}
		for _, t := range []string{c.uniquePrefix, chatPrefix} {
			fullHTML = strings.TrimSpace(strings.TrimPrefix(fullHTML, t))
		}
		i := strings.IndexAny(fullHTML, " \n\t")
		if i != -1 {
			fullHTML = fullHTML[i+1:]
		} else {
			fullHTML = ""
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

	return &Context{ctx, cctx, args, fullHTML, members, chatID, parsedDates}
}

func (c *Command) getChatID(ctx context.Context, msg *gotgbot.Message) (int64, error) {
	if msg.Chat.Type != "private" {
		return msg.Chat.Id, nil
	}
	if msg.From == nil {
		return 0, nil
	}
	return c.sessionService.GetActiveChat(ctx, msg.From.Id)
}

func (c *Command) findMembersByTitle(ctx context.Context, msg *gotgbot.Message, rest string) (string, []*model.ChatMember) {
	if rest == "" {
		return rest, nil
	}

	hasOtherUsers := msg.ReplyToMessage != nil
	if !hasOtherUsers {
		for _, e := range msg.GetEntities() {
			if e.Type == "mention" || e.Type == "text_mention" {
				hasOtherUsers = true
				break
			}
		}
	}
	if hasOtherUsers {
		return rest, nil
	}

	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return rest, nil
	}

	title := parts[0]
	chatID, err := c.getChatID(ctx, msg)
	if err != nil {
		return rest, nil
	}

	members, err := c.userService.GetByTag(ctx, chatID, title)
	if err != nil || len(members) == 0 {
		return rest, nil
	}

	res := make([]*model.ChatMember, 0, len(members))
	for _, m := range members {
		mm := m
		res = append(res, &mm)
	}

	newRest := strings.TrimSpace(strings.TrimPrefix(rest, title))
	return newRest, res
}

func (c *Command) Name() string {
	if len(c.commands) > 0 {
		return "command_" + c.commands[0]
	}
	return "unnamed_command"

}

func GetChatID(sessionService SessionService, ctx *ext.Context, cctx context.Context) (int64, error) {
	if ctx.EffectiveChat.Type != "private" {
		return ctx.EffectiveChat.Id, nil
	}

	targetID, err := sessionService.GetActiveChat(cctx, ctx.EffectiveUser.Id)
	if err != nil {
		return 0, err
	}

	return targetID, nil
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

	if len(rest) > 0 {
		strippedRest, members := c.findMembersByTitle(ctx, msg, rest)
		if len(members) > 0 {
			rest = strippedRest
		}
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

	for _, e := range msg.GetEntities() {
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

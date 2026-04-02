package command

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/go-faster/errors"
)

type UserProvider interface {
	GetUser(ctx context.Context, id int64) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
}

type ChatMemberProvider interface {
	GetChatMember(ctx context.Context, chatID, userId int64) (model.ChatMember, error)
	GetChatMemberByUsername(ctx context.Context, chatID int64, username string) (model.ChatMember, error)
}

type ChatProvider interface {
	GetChat(ctx context.Context, chatID int64) (model.Chat, error)
	GetCommandPermission(ctx context.Context, chatID int64, name string) (model.Status, error)
}

type SessionService interface {
	GetChat(ctx context.Context, userID int64) (model.Chat, error)
}

type Scope int

const (
	ScopeChat Scope = iota
	ScopeUser
)

type Command struct {
	name        string
	scope       Scope
	response    Response
	category    Category
	argRules    []ArgRule
	description string
	triggers    []string
	aliases     []string

	isDevCommand bool
	devID        int64

	requireTrigger bool
	requiredStatus model.Status

	userProvider       UserProvider
	chatMemberProvider ChatMemberProvider
	chatProvider       ChatProvider
	sessionService     SessionService

	dateParser *helpers.DateParser
}

func NewCommand(name string, response Response) *Command {
	return &Command{name: name, response: response, dateParser: helpers.NewDateParser()}
}

func (c *Command) SetProviders(userProvider UserProvider, memberProvider ChatMemberProvider, chatProvider ChatProvider, sessionService SessionService) *Command {
	c.userProvider = userProvider
	c.chatProvider = chatProvider
	c.sessionService = sessionService
	c.chatMemberProvider = memberProvider

	return c
}

func (c *Command) SetCategory(category Category) *Command {
	c.category = category

	return c
}

func (c *Command) Category() Category {
	return c.category
}

func (c *Command) SetScope(scope Scope) *Command {
	c.scope = scope

	return c
}

func (c *Command) SetArgRules(rules ...ArgRule) *Command {
	c.argRules = rules

	return c
}

func (c *Command) IsDevCommand() bool {
	return c.isDevCommand
}

func (c *Command) SetDevCommand(isDevCommand bool) *Command {
	c.isDevCommand = isDevCommand

	return c
}

func (c *Command) SetDevID(devID int64) *Command {
	c.devID = devID

	return c
}

func (c *Command) DevID() int64 {
	return c.devID
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Description() string {
	return c.description
}

func (c *Command) SetDescription(description string) *Command {
	c.description = description

	return c
}

func (c *Command) Triggers() []string {
	return c.triggers
}

func (c *Command) SetTriggers(triggers ...string) *Command {
	c.triggers = triggers

	return c
}

func (c *Command) AddTriggers(triggers ...string) *Command {
	c.triggers = append(c.triggers, triggers...)

	return c
}

func (c *Command) Aliases() []string {
	return c.aliases
}

func (c *Command) SetAliases(aliases ...string) *Command {
	c.aliases = aliases

	return c
}

const (
	EntityTypeMention     = "mention"
	EntityTypeTextMention = "text_mention"
	EntityTypeUrl         = "url"
)

func (c *Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	stdCtx := context.Background()
	handlerCtx := Context{stdContext: stdCtx}

	msg := ctx.Message
	if msg == nil || msg.ForwardOrigin != nil || msg.GetText() == "" {
		return false
	}

	text := msg.GetText()
	entities := msg.GetEntities()

	if c.scope == ScopeChat {
		chat, err := c.getChat(ctx, stdCtx)
		if err != nil {
			logger.L.Warn("get chat failed", "error", err)
			return false
		}
		handlerCtx.chat = &chat
		c.triggers = append(c.triggers, chat.CommandPrefix)
		c.requireTrigger = !chat.AllowPrefixless
		handlerCtx.requiredStatus = c.requiredStatus
		if s, err := c.chatProvider.GetCommandPermission(stdCtx, chat.ID, c.name); err == nil {
			c.requiredStatus = s
			handlerCtx.requiredStatus = s
		}
	}

	// command validation
	trigger := c.findTrigger(text)
	if c.requireTrigger && trigger == "" {
		return false
	}

	alias := c.findAlias(text, trigger, b.User.Username)
	if alias == "" {
		return false
	}

	// sender
	var senderMember *model.ChatMember
	if ctx.Data != nil && ctx.Data["_cached_sender"] != nil {
		if cached, ok := ctx.Data["_cached_sender"].(*model.ChatMember); ok {
			senderMember = cached
		}
	}
	if senderMember == nil {
		member, err := c.resolveMember(stdCtx, handlerCtx.chat, msg.From.Id)
		if err != nil {
			logger.L.Error("get chat member failed", "error", err)
			return false
		}
		senderMember = member
		if ctx.Data == nil {
			ctx.Data = make(map[string]interface{})
		}
		ctx.Data["_cached_sender"] = senderMember
	}
	handlerCtx.senderChatMember = senderMember

	// args validation
	textNoPrefix := strings.TrimSpace(trimPrefixIgnoreCase(text, trigger))
	textNoCommand := strings.TrimSpace(trimPrefixIgnoreCase(textNoPrefix, alias))
	if len(c.argRules) == 0 && textNoCommand != "" {
		return false
	}
	handlerCtx.RawArgs = textNoCommand
	runeOffset := len([]rune(text)) - len([]rune(textNoCommand))
	html := msg.OriginalHTML()
	if html == "" {
		html = msg.OriginalCaptionHTML()
	}
	handlerCtx.RawArgsHTML = sliceHTMLByTextOffset(html, runeOffset)

	for _, rule := range c.argRules {
		switch rule.Type {
		case ArgTypeOnlyUserSender:
			if isValidReply(msg.ReplyToMessage) {
				return false
			}
			members, _, err := c.extractMembersFromEntities(stdCtx, handlerCtx.chat, text, entities)
			if err != nil {
				logger.L.Error("failed to extract users", "error", err)
				return false
			}
			if len(members) > 0 {
				return false
			}

		case ArgTypeAnyUser, ArgTypeMentionedUser:
			if err := c.resolveUsers(stdCtx, &handlerCtx, msg, text, entities); err != nil {
				return false
			}

			if rule.Type == ArgTypeMentionedUser {
				totalUsers := handlerCtx.chatMembers
				if replyUser := handlerCtx.replyChatMember; replyUser != nil {
					totalUsers = append(totalUsers, *replyUser)
				}
				if len(totalUsers) < rule.Min {
					return false
				}
			}
		case ArgTypeNumber:
			parsed := 0
			for _, tok := range freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets) {
				num, err := strconv.Atoi(tok.text)
				if err != nil {
					continue
				}
				handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{tok.start, tok.end})
				handlerCtx.numbers = append(handlerCtx.numbers, num)
				parsed++
				if rule.Max != MaxAny && parsed >= rule.Max {
					break
				}
			}
			if parsed < rule.Min {
				return false
			}

		case ArgTypeDate:
			parsed := 0
			toks := freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets)

			for i := 0; i < len(toks); {
				matched := false
				for width := 3; width >= 1; width-- {
					if i+width > len(toks) {
						continue
					}
					words := make([]string, width)
					for k := 0; k < width; k++ {
						words[k] = toks[i+k].text
					}
					t, ok := c.dateParser.Parse(strings.Join(words, " "))
					if !ok {
						continue
					}
					for k := 0; k < width; k++ {
						handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{toks[i+k].start, toks[i+k].end})
					}
					handlerCtx.dates = append(handlerCtx.dates, t)
					parsed++
					i += width
					matched = true
					break
				}
				if !matched {
					i++
				}
				if rule.Max != MaxAny && parsed >= rule.Max {
					break
				}
			}
			if parsed < rule.Min {
				return false
			}

		case ArgTypeText:
			if rule.Variadic {
				var parts []string
				for _, tok := range freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets) {
					parts = append(parts, tok.text)
				}
				joined := strings.Join(parts, " ")
				if joined == "" && rule.Min > 0 {
					return false
				}
				if joined != "" {
					handlerCtx.texts = append(handlerCtx.texts, joined)
				}
			} else {
				parsed := 0
				for _, tok := range freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets) {
					handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{tok.start, tok.end})
					handlerCtx.texts = append(handlerCtx.texts, tok.text)
					parsed++
					if rule.Max != MaxAny && parsed >= rule.Max {
						break
					}
				}
				if parsed < rule.Min {
					return false
				}
			}
		}
	}
	ctx.Data["handlerCtx"] = handlerCtx

	return true
}

func (c *Command) resolveUsers(ctx context.Context, handlerCtx *Context, msg *gotgbot.Message, text string, entities []gotgbot.MessageEntity) error {
	// reply user
	if isValidReply(msg.ReplyToMessage) {
		replyMember, err := c.resolveMember(ctx, handlerCtx.chat, msg.ReplyToMessage.From.Id)
		if err != nil {
			logger.L.Error("resolve member failed", "error", err)
			return err
		}
		handlerCtx.replyChatMember = replyMember
	}

	// mentioned users
	mentionMembers, memberOffsets, err := c.extractMembersFromEntities(ctx, handlerCtx.chat, text, entities)
	if err != nil {
		logger.L.Error("failed to extract users from entities", "error", err)
		return err
	}
	handlerCtx.chatMembers = mentionMembers

	rawArgsByteOffset := strings.Index(text, handlerCtx.RawArgs)
	if rawArgsByteOffset < 0 {
		rawArgsByteOffset = 0
	}
	for _, o := range memberOffsets {
		start := o.Start - rawArgsByteOffset
		end := o.End - rawArgsByteOffset
		if start < 0 || end <= 0 {
			continue
		}
		handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{start, end})
	}

	return nil
}

func (c *Command) SetRequiredStatus(status model.Status) *Command {
	c.requiredStatus = status

	return c
}

func (c *Command) RequiredStatus() model.Status {
	return c.requiredStatus
}

func (c *Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	raw, ok := ctx.Data["handlerCtx"]
	if !ok {
		return errors.New("no handlerCtx")
	}

	handlerCtx := raw.(Context)
	handlerCtx.Context = ctx

	sender, err := handlerCtx.Sender()
	if err != nil {
		return err
	}

	if (!c.isDevCommand || sender.User.ID != c.devID) && !sender.StatusGranted(handlerCtx.requiredStatus) {
		return handlerCtx.Reply(b, fmt.Sprintf("[%d/%d] Недостаточно прав для выполнения команды", sender.Status, handlerCtx.requiredStatus), nil)
	}
	return c.response(b, &handlerCtx)
}

func (c *Command) resolveMember(ctx context.Context, chat *model.Chat, userID int64) (*model.ChatMember, error) {
	if c.scope == ScopeChat {
		if chat == nil {
			return nil, errors.New("chat cannot be nil")
		}

		member, err := c.chatMemberProvider.GetChatMember(ctx, chat.ID, userID)
		if err != nil {
			return nil, err
		}
		return &member, nil
	}

	user, err := c.userProvider.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.ChatMember{
		User: user,
	}, nil
}

func (c *Command) findTrigger(text string) string {
	for _, t := range c.triggers {
		if hasPrefixIgnoreCase(text, t) {
			return t
		}
	}
	return ""
}

func hasPrefixIgnoreCase(s, prefix string) bool {
	sRunes := []rune(s)
	prefixRunes := []rune(prefix)
	if len(prefixRunes) > len(sRunes) {
		return false
	}
	return strings.EqualFold(string(sRunes[:len(prefixRunes)]), prefix)
}

func trimPrefixIgnoreCase(s, prefix string) string {
	sRunes := []rune(s)
	prefixRunes := []rune(prefix)
	if len(prefixRunes) > len(sRunes) {
		return s
	}
	if strings.EqualFold(string(sRunes[:len(prefixRunes)]), prefix) {
		return string(sRunes[len(prefixRunes):])
	}
	return s
}

func (c *Command) findAlias(text, trigger, botUsername string) string {
	text = strings.TrimSpace(trimPrefixIgnoreCase(text, trigger))
	allCommands := append(c.aliases, c.name)
	for _, alias := range allCommands {
		if hasPrefixIgnoreCase(text, alias) {
			remaining := trimPrefixIgnoreCase(text, alias)

			if strings.HasPrefix(remaining, "@") {
				atPart := strings.SplitN(remaining, " ", 2)[0]
				if !strings.EqualFold(atPart, "@"+botUsername) {
					continue
				}
				remaining = strings.TrimPrefix(remaining, atPart)
			}

			remainingRunes := []rune(remaining)
			if len(remainingRunes) == 0 || isDelimiter(remainingRunes[0]) {
				return alias
			}
		}
	}
	return ""
}

func isDelimiter(r rune) bool {
	return unicode.IsSpace(r) ||
		strings.ContainsRune(".,!?;:()[]{}<>/\\\"'`", r)
}

func isValidReply(replyToMessage *gotgbot.Message) bool {
	return replyToMessage != nil &&
		replyToMessage.ForumTopicCreated == nil &&
		replyToMessage.From != nil &&
		!replyToMessage.From.IsBot &&
		!replyToMessage.IsAutomaticForward
}

func (c *Command) getChat(ctx *ext.Context, stdCtx context.Context) (model.Chat, error) {
	if ctx.Data != nil {
		if cached, ok := ctx.Data["_cached_chat"].(model.Chat); ok {
			return cached, nil
		}
	}

	msg := ctx.EffectiveMessage

	var chat model.Chat
	var err error

	if ctx.EffectiveChat.Type == gotgbot.ChatTypePrivate {
		chat, err = c.sessionService.GetChat(stdCtx, msg.From.Id)
		if err != nil {
			return model.Chat{}, errors.Wrap(err, "failed to get chat from private messages")
		}
	} else {
		chat, err = c.chatProvider.GetChat(stdCtx, msg.Chat.Id)
		if err != nil {
			return model.Chat{}, errors.Wrap(err, "failed to get chat from group")
		}
	}

	if ctx.Data == nil {
		ctx.Data = make(map[string]interface{})
	}
	ctx.Data["_cached_chat"] = chat
	return chat, nil
}

func (c *Command) extractMembersFromEntities(
	ctx context.Context,
	chat *model.Chat,
	text string,
	entities []gotgbot.MessageEntity,
) ([]model.ChatMember, []Offset, error) {

	var members []model.ChatMember
	var offsets []Offset

	for _, entity := range entities {
		extracted := extractEntity(text, entity)

		// Convert UTF-16 entity offset/length to byte offset in text.
		encoded := utf16.Encode([]rune(text))
		byteStart := len(string(utf16.Decode(encoded[:entity.Offset])))
		byteEnd := byteStart + len(string(utf16.Decode(encoded[entity.Offset:entity.Offset+entity.Length])))
		entityOffset := Offset{byteStart, byteEnd}

		switch entity.Type {

		case EntityTypeTextMention:
			member, err := c.resolveMember(ctx, chat, entity.User.Id)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, *member)
			offsets = append(offsets, entityOffset)

		case EntityTypeMention:
			username := parseUsernameFromMention(extracted)

			member, err := c.resolveMemberByUsername(ctx, chat, username)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, *member)
			offsets = append(offsets, entityOffset)

		case EntityTypeUrl:
			username := parseUsernameFromUrl(extracted)

			member, err := c.resolveMemberByUsername(ctx, chat, username)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, *member)
			offsets = append(offsets, entityOffset)
		}
	}

	return members, offsets, nil
}

func (c *Command) resolveMemberByUsername(
	ctx context.Context,
	chat *model.Chat,
	username string,
) (*model.ChatMember, error) {

	if c.scope == ScopeChat {
		if chat == nil {
			return nil, errors.New("chat cannot be nil")
		}

		member, err := c.chatMemberProvider.GetChatMemberByUsername(ctx, chat.ID, username)
		if err != nil {
			return nil, err
		}
		return &member, nil
	}

	user, err := c.userProvider.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &model.ChatMember{
		User: user,
	}, nil
}

func extractEntity(text string, e gotgbot.MessageEntity) string {
	encoded := utf16.Encode([]rune(text))
	slice := encoded[e.Offset : e.Offset+e.Length]
	return string(utf16.Decode(slice))
}

func parseUsernameFromUrl(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "t.me/")
	parts := strings.Split(url, "/")
	return strings.TrimPrefix(parts[0], "@")
}

func parseUsernameFromMention(mention string) string {
	return strings.TrimPrefix(mention, "@")
}

func sliceHTMLByTextOffset(html string, offset int) string {
	return string([]rune(html)[offset:])
}

// token is a whitespace-separated word with its byte offsets in the source string.
type token struct {
	text  string
	start int
	end   int
}

// isRangeUsed reports whether [start, end) overlaps with any used offset.
func isRangeUsed(start, end int, used []Offset) bool {
	for _, o := range used {
		if start < o.End && end > o.Start {
			return true
		}
	}
	return false
}

// freeTokens splits s into whitespace-separated tokens, returning only those
// whose byte range does not overlap with any entry in used.
func freeTokens(s string, used []Offset) []token {
	var tokens []token
	i := 0
	for i < len(s) {
		// skip whitespace
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\n' && s[j] != '\r' {
			j++
		}
		if !isRangeUsed(i, j, used) {
			tokens = append(tokens, token{text: s[i:j], start: i, end: j})
		}
		i = j
	}
	return tokens
}

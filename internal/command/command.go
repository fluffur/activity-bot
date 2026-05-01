package command

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"context"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

type UserProvider interface {
	GetUser(ctx context.Context, id int64) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
	EnsureUserExists(ctx context.Context, id int64, username, firstName, lastName string, isBot bool) (model.User, error)
}

type ChatMemberProvider interface {
	GetChatMember(ctx context.Context, chatID, userId int64) (model.ChatMember, error)
	GetChatMemberByUsername(ctx context.Context, chatID int64, username string) (model.ChatMember, error)
	EnsureMemberExists(ctx context.Context, chatID int64, userID int64, username, firstName, lastName, role string, isBot bool) (model.ChatMember, error)
	FindChatMembersByTag(ctx context.Context, chatID int64, tag string) ([]model.ChatMember, error)
}

type ChatProvider interface {
	GetChat(ctx context.Context, chatID int64) (model.Chat, error)
	EnsureChatExists(ctx context.Context, chatID int64, title string) (model.Chat, error)
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
	name         string
	scope        Scope
	response     Response
	category     Category
	argRules     []ArgRule
	description  string
	prefixes     []string
	aliases      []string
	isDevCommand bool
	devID        int64

	requirePrefix bool
	middlewares   []Middleware

	requiredStatus model.Status

	checkStatusDisabled bool

	userProvider       UserProvider
	chatMemberProvider ChatMemberProvider
	chatProvider       ChatProvider
	sessionService     SessionService

	important  bool
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

func (c *Command) Scope() Scope {
	return c.scope
}

func (c *Command) SetArgRules(rules ...ArgRule) *Command {
	c.argRules = rules

	return c
}

func (c *Command) ArgRules() []ArgRule {
	return c.argRules
}

func (c *Command) SetImportant(flag bool) *Command {
	c.important = flag

	return c
}

func (c *Command) Important() bool {
	return c.important
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
	return c.prefixes
}

func (c *Command) SetPrefixes(prefixes ...string) *Command {
	c.prefixes = prefixes

	return c
}

func (c *Command) AddPrefixes(prefixes ...string) *Command {
	c.prefixes = append(c.prefixes, prefixes...)

	return c
}

func (c *Command) Aliases() []string {
	return c.aliases
}

func (c *Command) SetAliases(aliases ...string) *Command {
	c.aliases = aliases

	return c
}

func (c *Command) WithMiddlewares(middlewares ...Middleware) *Command {
	c.middlewares = middlewares

	return c
}

func (c *Command) findPrefixInList(text string, prefixes []string) string {
	for _, t := range prefixes {
		if hasPrefixIgnoreCase(text, t) {
			return t
		}
	}
	return ""
}

func (c *Command) CheckUpdate(ctx *ext.Context, u *ext.Update) error {
	handlerCtx := Context{Context: ctx, Command: c}
	m := u.EffectiveMessage
	if m == nil || m.Text == "" {
		return dispatcher.ContinueGroups
	}
	if m.EditDate != 0 {
		return dispatcher.ContinueGroups
	}

	text := m.Text
	entities := m.Entities

	// command validation
	currentPrefixes := make([]string, len(c.prefixes))
	copy(currentPrefixes, c.prefixes)
	requirePrefix := c.requirePrefix

	if c.scope == ScopeChat {
		chat, err := c.getChat(ctx, u)
		if err != nil {
			return errors.Wrap(err, "get chat failed")
		}
		handlerCtx.chat = &chat

		if chat.CommandPrefix != "" {
			currentPrefixes = append(currentPrefixes, chat.CommandPrefix)
		}
		requirePrefix = !chat.AllowPrefixless
		handlerCtx.requiredStatus = c.requiredStatus
		if s, err := c.chatProvider.GetCommandPermission(ctx.Context, chat.ID, c.name); err == nil {
			handlerCtx.requiredStatus = s
		}
	}

	prefix := c.findPrefixInList(text, currentPrefixes)
	alias := c.findAlias(text, prefix, ctx.Self.Username)
	if alias == "" {
		return dispatcher.ContinueGroups
	}

	if requirePrefix && prefix == "" {
		return dispatcher.ContinueGroups
	}

	// sender
	member, err := c.resolveMember(ctx, handlerCtx.chat, u.EffectiveUser())
	if err != nil {
		logger.L.Error("resolve sender failed", "error", err)
	}
	handlerCtx.senderChatMember = member

	// args validation
	textNoPrefix := strings.TrimSpace(trimPrefixIgnoreCase(text, prefix))
	textNoCommand := strings.TrimSpace(trimPrefixIgnoreCase(trimPrefixIgnoreCase(textNoPrefix, alias), "@"+ctx.Self.Username))

	if len(c.argRules) == 0 && textNoCommand != "" {
		return dispatcher.ContinueGroups
	}

	handlerCtx.RawArgs = textNoCommand
	if textNoCommand != "" {
		trimmedText := trimPrefixIgnoreCase(text, prefix)
		trimmedText = strings.TrimLeft(trimmedText, " \t\n\r")
		trimmedText = trimPrefixIgnoreCase(trimmedText, alias)
		if strings.HasPrefix(trimmedText, "@") {
			atPart := strings.SplitN(trimmedText, " ", 2)[0]
			if strings.EqualFold(atPart, "@"+ctx.Self.Username) {
				trimmedText = trimmedText[len(atPart):]
			}
		}
		trimmedText = strings.TrimLeft(trimmedText, " \t\n\r")

		fullUTF16 := utf16.Encode([]rune(text))
		argUTF16 := utf16.Encode([]rune(trimmedText))
		offsetUTF16 := len(fullUTF16) - len(argUTF16)
		handlerCtx.RawArgsEntities = shiftEntities(entities, offsetUTF16)
	}

	for _, rule := range c.argRules {
		switch rule.Type {
		case ArgTypeOnlyUserSender:
			if _, ok := getReplyToMessageID(m); ok {
				return dispatcher.ContinueGroups
			}
			members, _, err := c.extractMembersFromEntities(ctx, handlerCtx.chat, text, entities)
			if err != nil {
				return errors.Wrap(err, "failed to extract users")
			}
			if len(members) > 0 {
				return dispatcher.ContinueGroups
			}

		case ArgTypeAnyUser, ArgTypeMentionedUser:
			if err := c.resolveUsers(ctx, &handlerCtx, m, text, entities); err != nil {
				return err
			}
			if c.scope == ScopeChat && handlerCtx.chat != nil && len(handlerCtx.chatMembers) == 0 && handlerCtx.replyChatMember == nil {
				toks := freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets)
				matched := false
				for i := 0; i < len(toks) && !matched; {
					for width := 3; width >= 1; width-- {
						if i+width > len(toks) {
							continue
						}
						words := make([]string, width)
						for k := 0; k < width; k++ {
							words[k] = toks[i+k].text
						}
						tag := strings.Join(words, " ")
						if len([]rune(tag)) <= 16 {
							members, err := c.chatMemberProvider.FindChatMembersByTag(ctx.Context, handlerCtx.chat.ID, tag)
							if err == nil && len(members) > 0 {
								handlerCtx.chatMembers = append(handlerCtx.chatMembers, members...)
								for k := 0; k < width; k++ {
									handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{toks[i+k].start, toks[i+k].end})
								}
								matched = true
								break
							}
						}
					}
					if !matched {
						i++
					}
				}
			}

			if rule.Type == ArgTypeMentionedUser {
				totalUsers := handlerCtx.chatMembers
				if replyUser := handlerCtx.replyChatMember; replyUser != nil {
					totalUsers = append(totalUsers, *replyUser)
				}
				if len(totalUsers) < rule.Min {
					return dispatcher.ContinueGroups
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
				return dispatcher.ContinueGroups
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
				return dispatcher.ContinueGroups
			}

		case ArgTypeText:
			if rule.Variadic {
				var parts []string
				for _, tok := range freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets) {
					parts = append(parts, tok.text)
				}
				joined := strings.Join(parts, " ")
				if joined == "" && rule.Min > 0 {
					return dispatcher.ContinueGroups
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
					return dispatcher.ContinueGroups
				}
			}
		}
	}

	if !c.checkStatusDisabled && handlerCtx.senderChatMember != nil && !handlerCtx.senderChatMember.StatusGranted(handlerCtx.requiredStatus) {

		if err := handlerCtx.ReplyOnly(u, options.WithText(fmt.Sprintf("Требуются права: %s", handlerCtx.requiredStatus))); err != nil {
			logger.L.Error("reply status", "error", err)
		}

		return dispatcher.EndGroups
	}

	for _, middleware := range c.middlewares {
		if err := middleware.CheckUpdate(&handlerCtx, u); err != nil {
			if errors.Is(err, ErrStop) {
				return dispatcher.EndGroups
			}
			logger.L.Error("middleware", "error", err)
			return dispatcher.EndGroups
		}
	}

	err = c.response(&handlerCtx, u)

	if err != nil {
		logger.L.Error("response", "error", err)
	}
	log.Println("update handled", c.name)
	return dispatcher.EndGroups
}

func (c *Command) resolveUsers(ctx *ext.Context, handlerCtx *Context, msg *types.Message, text string, entities []tg.MessageEntityClass) error {
	// reply user
	if msgID, ok := getReplyToMessageID(msg); ok {
		messages, err := ctx.GetMessages(handlerCtx.chat.ID, []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}})
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return errors.New("no reply message")
		}
		user, err := c.extractUserFromMessage(ctx, messages[0])
		if err != nil {
			return errors.Wrap(err, "extract user from reply failed")
		}
		replyMember, err := c.resolveMember(ctx, handlerCtx.chat, user)
		if err != nil {
			return errors.Wrap(err, "resolve reply failed")
		}
		handlerCtx.replyChatMember = replyMember
	}

	// mentioned users
	mentionMembers, memberOffsets, err := c.extractMembersFromEntities(ctx, handlerCtx.chat, text, entities)
	if err != nil {
		return errors.Wrap(err, "extract members failed")
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

func (c *Command) extractUserFromMessage(
	ctx *ext.Context,
	msg tg.MessageClass,
) (*tg.User, error) {

	var userID int64

	switch m := msg.(type) {

	case *tg.Message:
		fromID, ok := m.GetFromID()
		if !ok {
			return nil, errors.New("message has no FromID")
		}

		peerUser, ok := fromID.(*tg.PeerUser)
		if !ok {
			return nil, errors.New("FromID is not PeerUser")
		}

		userID = peerUser.UserID

	case *tg.MessageService:

		switch action := m.Action.(type) {
		case *tg.MessageActionChatJoinedByLink:
			peerUser, ok := m.FromID.(*tg.PeerUser)

			if !ok {
				return nil, errors.New("PeerID is not PeerUser")
			}
			userID = peerUser.UserID

		case *tg.MessageActionChatAddUser:
			if len(action.Users) == 0 {
				return nil, errors.New("no users in ChatAddUser")
			}
			userID = action.Users[0]

		case *tg.MessageActionChatDeleteUser:
			userID = action.UserID

		default:
			return nil, errors.New("unsupported service action")
		}

	default:
		slog.Error("unsupported msg type", "msg", m)
		return nil, errors.New("unsupported message type")
	}

	inputPeer, err := ctx.ResolveInputPeerById(userID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve input peer failed")
	}

	pUser, ok := inputPeer.(*tg.InputPeerUser)
	if !ok {
		return nil, errors.New("not InputPeerUser")
	}

	users, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{
		&tg.InputUser{
			UserID:     pUser.UserID,
			AccessHash: pUser.AccessHash,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	user, ok := users[0].(*tg.User)
	if !ok {
		return nil, errors.New("not tg.User")
	}

	return user, nil
}

func (c *Command) SetRequiredStatus(status model.Status) *Command {
	c.requiredStatus = status

	return c
}

func (c *Command) DisableCheckStatus() *Command {
	c.checkStatusDisabled = true

	return c
}

func (c *Command) RequiredStatus() model.Status {
	return c.requiredStatus
}

func (c *Command) resolveMember(ctx *ext.Context, chat *model.Chat, u any) (*model.ChatMember, error) {
	var userObj *tg.User

	switch val := u.(type) {
	case *tg.User:
		userObj = val
	case int64:
		inputPeer, err := ctx.ResolveInputPeerById(val)
		if err != nil {
			return nil, fmt.Errorf("resolve peer failed: %w", err)
		}
		pUser, ok := inputPeer.(*tg.InputPeerUser)
		if !ok {
			return nil, fmt.Errorf("peer %d is not a user", val)
		}
		uSlice, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUser{
			UserID:     pUser.UserID,
			AccessHash: pUser.AccessHash,
		}})
		if err != nil {
			return nil, fmt.Errorf("fetch user failed: %w", err)
		}
		if len(uSlice) == 0 {
			return nil, fmt.Errorf("user %d not found", val)
		}
		userObj, ok = uSlice[0].(*tg.User)
		if !ok {
			return nil, fmt.Errorf("peer %d is not a user", val)
		}
	default:
		return nil, fmt.Errorf("unsupported type for resolveMember: %T", u)
	}

	if c.scope == ScopeChat {
		if chat == nil {
			return nil, errors.New("chat cannot be nil")
		}

		member, err := c.chatMemberProvider.EnsureMemberExists(
			ctx.Context,
			chat.ID,
			userObj.GetID(),
			userObj.Username,
			userObj.FirstName,
			userObj.LastName,
			"",
			userObj.Bot,
		)
		if err != nil {
			return nil, err
		}
		return &member, nil
	}

	user, err := c.userProvider.EnsureUserExists(
		ctx.Context,
		userObj.GetID(),
		userObj.Username,
		userObj.FirstName,
		userObj.LastName,
		userObj.Bot,
	)
	if err != nil {
		return nil, err
	}

	return &model.ChatMember{
		User: user,
	}, nil
}

func (c *Command) findPrefix(text string) string {
	for _, t := range c.prefixes {
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

func getReplyToMessageID(msg *types.Message) (int, bool) {
	replyTo, ok := msg.GetReplyTo()
	if !ok {
		return 0, false
	}
	header, ok := replyTo.(*tg.MessageReplyHeader)
	if !ok {
		return 0, false
	}

	return header.GetReplyToMsgID()
}

func (c *Command) getChat(ctx *ext.Context, u *ext.Update) (model.Chat, error) {
	ec := u.EffectiveChat()

	switch chat := ec.(type) {

	case *types.User:
		result, err := c.sessionService.GetChat(ctx.Context, chat.GetID())
		if err != nil {
			return model.Chat{}, errors.Wrap(err, "failed to get private chat")
		}
		return result, nil

	case *types.Chat, *types.Channel:
		var title string

		switch ch := chat.(type) {
		case *types.Chat:
			title = ch.Title
		case *types.Channel:
			title = ch.Title
		}

		result, err := c.chatProvider.EnsureChatExists(
			ctx.Context,
			chat.GetID(),
			title,
		)
		if err != nil {
			return model.Chat{}, errors.Wrap(err, "failed to ensure chat")
		}

		return result, nil

	default:
		return model.Chat{}, errors.New("unsupported chat type")
	}
}

func (c *Command) extractMembersFromEntities(
	ctx *ext.Context,
	chat *model.Chat,
	text string,
	entities []tg.MessageEntityClass,
) ([]model.ChatMember, []Offset, error) {

	var members []model.ChatMember
	var offsets []Offset

	for _, entity := range entities {
		extracted := extractEntity(text, entity)

		encoded := utf16.Encode([]rune(text))
		byteOffset := entity.GetOffset()
		byteLength := entity.GetLength()
		byteStart := len(string(utf16.Decode(encoded[:byteOffset])))
		byteEnd := byteStart + len(string(utf16.Decode(encoded[byteOffset:byteOffset+byteLength])))
		entityOffset := Offset{byteStart, byteEnd}

		switch e := entity.(type) {

		case *tg.MessageEntityMentionName:
			member, err := c.resolveMember(ctx, chat, e.UserID)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, *member)
			offsets = append(offsets, entityOffset)

		case *tg.MessageEntityMention:
			username := parseUsernameFromMention(extracted)

			member, err := c.resolveMemberByUsername(ctx, chat, username)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, *member)
			offsets = append(offsets, entityOffset)

		case *tg.MessageEntityTextURL:
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

func extractEntity(text string, e tg.MessageEntityClass) string {
	encoded := utf16.Encode([]rune(text))
	slice := encoded[e.GetOffset() : e.GetOffset()+e.GetLength()]
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

func shiftEntities(entities []tg.MessageEntityClass, offset int) []tg.MessageEntityClass {
	var result []tg.MessageEntityClass
	for _, e := range entities {
		v := reflect.ValueOf(e)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			continue
		}

		offsetField := v.FieldByName("Offset")
		lengthField := v.FieldByName("Length")
		if !offsetField.IsValid() || !lengthField.IsValid() {
			continue
		}

		originalOffset := int(offsetField.Int())
		originalLength := int(lengthField.Int())
		end := originalOffset + originalLength

		if end <= offset {
			continue
		}

		newOffset := originalOffset - offset
		newLength := originalLength

		if newOffset < 0 {
			newLength += newOffset
			newOffset = 0
		}

		newEntityPtr := reflect.New(v.Type())
		newEntity := newEntityPtr.Elem()
		newEntity.Set(v)

		newEntity.FieldByName("Offset").SetInt(int64(newOffset))
		newEntity.FieldByName("Length").SetInt(int64(newLength))

		result = append(result, newEntityPtr.Interface().(tg.MessageEntityClass))
	}
	return result
}

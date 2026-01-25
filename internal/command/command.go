package command

import (
	"activity-bot/internal/common"
	"activity-bot/internal/model"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Builder struct {
	userService  common.UserService
	adminService common.AdminService
}

func NewBuilder(userService common.UserService, adminService common.AdminService) *Builder {
	return &Builder{userService, adminService}
}

func (b *Builder) New(c string, r Response, aliases ...string) Command {
	return NewCommand(c, r, b.userService, b.adminService, aliases...)
}

type Command struct {
	Command          string
	Triggers         []string
	Aliases          []string
	Response         Response
	MaxArgs          int
	requireAdmin     bool
	requireCreator   bool
	allowArgs        bool
	requireTriggers  bool
	fallbackToSender bool
	onlyGroups       bool
	userService      common.UserService
	adminService     common.AdminService
}

func NewCommand(c string, r Response, userService common.UserService, adminService common.AdminService, aliases ...string) Command {
	return Command{
		Command:          strings.ToLower(c),
		Triggers:         []string{"/", "!", ".", ""},
		Aliases:          aliases,
		Response:         r,
		requireTriggers:  true,
		fallbackToSender: false,

		userService:  userService,
		adminService: adminService,
	}
}

func (c Command) OnlyGroups() Command {
	c.onlyGroups = true
	return c
}

func (c Command) RequireTriggers(require bool) Command {
	c.requireTriggers = require
	return c
}

func (c Command) RequireAdmin() Command {
	c.requireAdmin = true
	return c
}

func (c Command) RequireCreator() Command {
	c.requireCreator = true
	return c
}

func (c Command) FallbackToSender() Command {
	c.fallbackToSender = true
	return c
}

func (c Command) AllowArgs() Command {
	c.allowArgs = true
	return c
}

func (c Command) SetMaxArgs(maxArgs int) Command {
	c.MaxArgs = maxArgs
	return c
}

func (c Command) SetTriggers(triggers ...string) Command {
	c.Triggers = triggers
	return c
}

func (c Command) SetAliases(aliases ...string) Command {
	c.Aliases = aliases
	return c
}

func (c Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if c.onlyGroups && ctx.EffectiveChat.Type == "private" {
		return false
	}

	if ctx.Message != nil {
		if ctx.Message.GetText() == "" {
			return false
		}
		return c.checkMessage(b, ctx.Message)
	}

	return false
}
func (c Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	if c.requireCreator {
		if !common.IsSenderCreator(b, ctx) {
			_, err := ctx.EffectiveMessage.Reply(b, "Только создатель может выполнить эту команду", nil)
			return err
		}
	}

	if c.requireAdmin {
		if !common.IsSenderAdmin(b, ctx, c.adminService) {
			_, err := ctx.EffectiveMessage.Reply(b, "Только создатель и администраторы могут выполнить эту команду", nil)
			return err
		}
	}

	return c.Response(b, ctx, c.parseArgs(b, ctx))
}

func (c Command) ensureUser(u *gotgbot.User) (model.User, error) {
	return c.userService.EnsureUserExists(u.Id, u.Username, u.FirstName, u.LastName)

}
func (c Command) parseArgs(b *gotgbot.Bot, ctx *ext.Context) *Context {
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
	for _, t := range c.Triggers {
		for _, cmd := range append([]string{c.Command}, c.Aliases...) {
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
	if c.MaxArgs > 0 && len(words) > c.MaxArgs {
		last := strings.Join(words[c.MaxArgs-1:], " ")
		words = append(words[:c.MaxArgs-1], last)
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

func (c Command) Name() string {
	return "command_" + c.Command
}

func (c Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	text := strings.ToLower(removeMentions(msg))
	if text == "" {
		return false
	}

	for _, trigger := range c.Triggers {
		if !strings.HasPrefix(text, trigger) {
			continue
		}

		for _, cName := range append([]string{c.Command}, c.Aliases...) {
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

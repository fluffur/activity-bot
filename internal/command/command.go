package command

import (
	"activity-bot/internal/model"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type UserService interface {
	GetUserByUsername(username string) (model.User, error)
	EnsureUserExists(id int64, username, firstName, lastName string) (model.User, error)
}

type Builder struct {
	userService UserService
}

func NewBuilder(us UserService) *Builder {
	return &Builder{
		userService: us,
	}
}

func (b *Builder) NewCommand(c string, r Response, aliases ...string) Command {
	return NewCommand(c, r, b.userService, aliases...)
}

type Command struct {
	Command     string
	Triggers    []rune
	Aliases     []string
	Response    Response
	MaxArgs     int
	userService UserService
}

func NewCommand(c string, r Response, service UserService, aliases ...string) Command {
	return Command{
		Command:     strings.ToLower(c),
		Triggers:    []rune("/!."),
		Aliases:     aliases,
		Response:    r,
		userService: service,
	}
}

func (c Command) SetMaxArgs(maxArgs int) Command {
	c.MaxArgs = maxArgs
	return c
}

func (c Command) SetTriggers(triggers []rune) Command {
	c.Triggers = triggers
	return c
}

func (c Command) SetAliases(aliases ...string) Command {
	c.Aliases = aliases
	return c
}

func (c Command) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	if ctx.Message != nil {
		if ctx.Message.GetText() == "" {
			return false
		}
		return c.checkMessage(b, ctx.Message)
	}

	return false
}
func (c Command) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	args, user := c.parseArgs(ctx.Message)

	return c.Response(b, ctx, user, args)
}

func (c Command) ensureUser(u *gotgbot.User) (model.User, error) {
	return c.userService.EnsureUserExists(u.Id, u.Username, u.FirstName, u.LastName)

}

func (c Command) parseArgs(msg *gotgbot.Message) ([]string, *model.User) {
	text := msg.GetText()
	textRunes := []rune(text)

	var targetUser *model.User

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		u, err := c.ensureUser(msg.ReplyToMessage.From)
		if err == nil {
			targetUser = &u
		} else {
			log.Println("Failed to ensure user", err)
		}
	}

	restRunes := textRunes
	for _, e := range msg.Entities {
		start := int(e.Offset)
		end := start + int(e.Length)

		if e.Type == "text_mention" && e.User != nil && targetUser == nil {
			u, err := c.ensureUser(e.User)
			if err == nil {
				targetUser = &u
			} else {
				log.Println("Failed to ensure user", err)

			}
		}
		if (e.Type == "mention") && targetUser == nil {
			username := string(restRunes[start+1 : end])
			u, err := c.userService.GetUserByUsername(username)
			if err == nil {
				targetUser = &u
			} else {
				log.Println("Failed to get user", err)

			}
		}

		if start >= 0 && end <= len(restRunes) {
			restRunes = append(restRunes[:start], restRunes[end:]...)
		}
	}
	rest := strings.TrimSpace(string(restRunes))

	commands := append([]string{c.Command}, c.Aliases...)
	for _, t := range c.Triggers {
		for _, cmd := range commands {
			fullCmd := string(t) + strings.ToLower(cmd)
			if strings.HasPrefix(strings.ToLower(rest), fullCmd) {
				rest = strings.TrimSpace(string([]rune(rest)[len([]rune(fullCmd)):]))

				// 4. делим на аргументы
				if c.MaxArgs <= 0 {
					return strings.Fields(rest), targetUser
				}
				words := strings.Fields(rest)
				if len(words) <= c.MaxArgs {
					return []string{rest}, targetUser
				}
				args := make([]string, 0, c.MaxArgs)
				for i := 0; i < c.MaxArgs-1; i++ {
					args = append(args, words[i])
				}
				last := strings.Join(words[c.MaxArgs-1:], " ")
				args = append(args, last)
				return args, targetUser
			}
		}
	}

	return []string{}, targetUser
}

func (c Command) Name() string {
	return "command_" + c.Command
}

func (c Command) checkMessage(b *gotgbot.Bot, msg *gotgbot.Message) bool {
	ents := msg.GetEntities()
	if len(ents) != 0 && ents[0].Offset == 0 && ents[0].Type != "bot_command" {
		return false
	}

	text := msg.GetText()

	var cmd string
	for _, t := range c.Triggers {
		if r, _ := utf8.DecodeRuneInString(text); r != t {
			continue
		}

		split := strings.Split(strings.ToLower(strings.Fields(text)[0]), "@")
		if len(split) > 1 && split[1] != strings.ToLower(b.User.Username) {
			return false
		}
		cmd = split[0][1:]
		break
	}
	if cmd == "" {
		return false
	}

	if cmd == c.Command {
		return true
	}

	for _, alias := range c.Aliases {
		if cmd == strings.ToLower(alias) {
			return true
		}
	}

	return false
}

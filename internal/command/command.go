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
	return c.Response(b, ctx, c.parseArgs(ctx.Message))
}

func (c Command) ensureUser(u *gotgbot.User) (model.User, error) {
	return c.userService.EnsureUserExists(u.Id, u.Username, u.FirstName, u.LastName)

}
func (c Command) parseArgs(msg *gotgbot.Message) *Context {
	text := msg.GetText()
	textRunes := []rune(text)

	usersMap := make(map[int64]*model.User)
	removeRanges := make([][2]int, 0)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
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

	if len(removeRanges) > 0 {
		for i := len(removeRanges) - 1; i >= 0; i-- {
			r := removeRanges[i]
			textRunes = append(textRunes[:r[0]], textRunes[r[1]:]...)
		}
	}

	rest := strings.TrimSpace(string(textRunes))

commandsLoop:
	for _, t := range c.Triggers {
		for _, cmd := range append([]string{c.Command}, c.Aliases...) {
			fullCmd := string(t) + strings.ToLower(cmd)
			if strings.HasPrefix(strings.ToLower(rest), fullCmd) {
				rest = strings.TrimSpace(rest[len(fullCmd):])
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

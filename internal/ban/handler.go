package ban

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

const CategoryBan command.Category = "ban"

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"banuser",
			h.BanUser,
			"Заблокировать пользователя глобально",
			CategoryBan,
			option.WithAliases("гбан"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
			option.OnlyDev(),
		),
		action.NewCommand(
			"unbanuser",
			h.UnbanUser,
			"Разблокировать пользователя глобально",
			CategoryBan,
			option.WithAliases("гразбан"),
			option.WithRules(
				rule.User().Optional(),
				rule.Number().Optional(),
			),
			option.OnlyDev(),
		),
		action.NewCommand(
			"banchat",
			h.BanChat,
			"Заблокировать чат глобально",
			CategoryBan,
			option.WithAliases("банчат"),
			option.WithRules(
				rule.Number(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
			option.OnlyDev(),
		),
		action.NewCommand(
			"unbanchat",
			h.UnbanChat,
			"Разблокировать чат глобально",
			CategoryBan,
			option.WithAliases("анбанчат"),
			option.WithRules(rule.Number()),
			option.OnlyDev(),
		),
	}
}

func (h *Handler) BanUser(c *botapi.Context) error {
	args := cctx.MustArgs(c)

	user, ok := args.User()
	if !ok {
		return nil
	}

	expiresAt, _ := args.Until()

	reason, _ := args.Text()

	if err := h.repo.BanUser(
		c,
		user.ID(),
		reason,
		expiresAt,
	); err != nil {
		return fmt.Errorf("ban user: %w", err)
	}

	_, err := c.Reply(
		fmt.Sprintf(
			"Пользователь %s заблокирован глобально%s.",
			tghtml.UserMention(user.ID(), user.Name("unknown")),
			formatBanExpiration(expiresAt),
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) UnbanUser(c *botapi.Context) error {
	args := cctx.MustArgs(c)

	var userID int64

	if user, ok := args.User(); ok {
		userID = user.ID()
	} else if id, ok := args.Number(); ok {
		userID = id
	} else {
		return nil
	}

	if err := h.repo.UnbanUser(c, userID); err != nil {
		return fmt.Errorf("unban user: %w", err)
	}

	_, err := c.Reply(
		fmt.Sprintf(
			"Пользователь %d разблокирован.",
			userID,
		),
	)

	return err
}

func (h *Handler) BanChat(c *botapi.Context) error {
	args := cctx.MustArgs(c)

	chatID, ok := args.Number()
	if !ok {
		return nil
	}

	expiresAt, _ := args.Until()
	reason, _ := args.Text()

	if err := h.repo.BanChat(
		c,
		chatID,
		reason,
		expiresAt,
	); err != nil {
		return fmt.Errorf("ban chat: %w", err)
	}

	_, err := c.Reply(
		fmt.Sprintf(
			"Чат <code>%d</code> заблокирован глобально%s.",
			chatID,
			formatBanExpiration(expiresAt),
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) UnbanChat(c *botapi.Context) error {
	args := cctx.MustArgs(c)

	chatID, ok := args.Number()
	if !ok {
		return nil
	}

	if err := h.repo.UnbanChat(c, chatID); err != nil {
		return fmt.Errorf("unban chat: %w", err)
	}

	_, err := c.Reply(
		fmt.Sprintf(
			"Чат <code>%d</code> разблокирован.",
			chatID,
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func formatBanExpiration(t time.Time) string {
	if t.IsZero() {
		return " навсегда"
	}

	return fmt.Sprintf(
		" до %s",
		t.Format("02.01.2006 15:04"),
	)
}

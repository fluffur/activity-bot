package reward

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gotd/botapi"
)

const CategoryRewards command.Category = "rewards"

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"reward",
			h.Reward,
			i18n.Cmd.Reward.Desc,
			CategoryRewards,
			option.WithRules(rule.User(), rule.Number().Optional(), rule.Text()),
			option.WithAliases("наградить"),
			option.WithPermission(permission.StatusModerator),
		),
		action.NewCommand(
			"unreward",
			h.UnReward,
			i18n.Cmd.Unreward.Desc,
			CategoryRewards,
			option.WithRules(rule.User().Optional(), rule.Number()),
			option.WithAliases("снять награду"),
			option.WithPermission(permission.StatusModerator),
			option.IgnorePermissionCheck(),
		),
		action.NewCommand(
			"rewards",
			h.ListRewards,
			i18n.Cmd.ListRewards.Desc,
			CategoryRewards,
			option.WithRules(rule.User().Optional()),
			option.WithAliases("награды"),
		),
	}
}

func (h *Handler) Reward(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	rank, ok := args.Number()
	if !ok {
		rank = 0
	}
	if rank < 0 || rank > 9 {
		return fmt.Errorf("reward: invalid rank")
	}

	reason, ok := args.Text()
	reason = strings.TrimSpace(reason)
	if !ok || utf8.RuneCountInString(reason) > 128 {
		return fmt.Errorf("reward: invalid reason")
	}
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	target, ok := args.User()
	if !ok {
		return fmt.Errorf("reward: invalid user")
	}
	if target.ID() == cm.ID() {
		return nil
	}

	if err := h.repo.AddReward(c, NewReward(ch.ID, target.ID(), cm.ID(), int16(rank), reason)); err != nil {
		return fmt.Errorf("reward: %w", err)
	}
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.Reward.Success, nil))
	return err
}

func (h *Handler) UnReward(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	n, ok := args.Number()
	if !ok || n <= 0 {
		return fmt.Errorf("unreward: no number")
	}
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	target, ok := args.User()
	if !ok {
		target = cm
	}

	if !cm.Permitted(cctx.MustPermission(c)) && cm.ID() != target.ID() {
		return fmt.Errorf("reward: not allowed to unreward")
	}

	rws, err := h.repo.ListRewards(c, ch.ID, target.ID())
	if err != nil {
		return fmt.Errorf("unreward: list rewards: %w", err)
	}
	index := int(n) - 1

	if index < 0 || index >= len(rws) {
		return fmt.Errorf("unreward: invalid reward number")
	}

	rwToRemove := rws[index]

	if err := h.repo.RemoveReward(c, rwToRemove.ID); err != nil {
		return fmt.Errorf("unreward: %w", err)
	}
	loc := cctx.MustLocalizer(c)

	_, err = c.Reply(loc.T(i18n.Cmd.Unreward.Success, nil))
	return err
}

func (h *Handler) ListRewards(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	u, ok := cctx.MustArgs(c).User()
	if ok {
		cm = u
	}
	rws, err := h.repo.ListRewards(c, ch.ID, cm.ID())
	if err != nil {
		return fmt.Errorf("unreward: list rewards: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	var b strings.Builder

	title := loc.T(i18n.Cmd.ListRewards.Title, i18n.CmdListRewardsTitleData{
		User: tghtml.MemberLink(loc, ch, cm),
	})

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString("<blockquote expandable>")
	if len(rws) == 0 {
		b.WriteString(loc.T(i18n.Cmd.ListRewards.NoRewards, nil))
	}
	given := loc.T(i18n.Cmd.ListRewards.Given, nil)
	for i, rw := range rws {
		b.WriteString(fmt.Sprintf("%d. %s %s",
			i+1,
			tghtml.Bold(tghtml.Escape(rw.Reason)),
			RankEmoji(rw.Rank),
		))

		b.WriteString("\n    ")
		b.WriteString(given)
		b.WriteString(" ")
		b.WriteString(tghtml.DefaultDateTime(rw.CreatedAt))

		if i != len(rws)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("</blockquote>")

	_, err = c.Reply(b.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)
	return err
}

func RankEmoji(rank int16) string {
	switch rank {
	case 0:
		return "❤️"
	case 1:
		return "⭐"
	case 2:
		return "🌟"
	case 3:
		return "💎"
	case 4:
		return "🔥"
	case 5:
		return "✨"
	case 6:
		return "🌸"
	case 7:
		return "👑"
	case 8:
		return "💠"
	case 9:
		return "🏆"
	default:
		return "❤️"
	}
}

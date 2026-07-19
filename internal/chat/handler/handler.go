package handler

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

const CategoryChat command.Category = "chat"

type Handler struct {
	service *chat.Service
}

func NewHandler(service *chat.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"newbie",
			h.ShowNewbieThreshold,
			i18n.Cmd.Chat.ShowNewbieThreshold.Desc,
			CategoryChat,
			option.WithAliases("новички"),
		),

		action.NewCommand(
			"setnewbie",
			h.SetNewbieThreshold,
			i18n.Cmd.Chat.SetNewbieThreshold.Desc,
			CategoryChat,
			option.WithAliases("новички"),
			option.WithRules(
				rule.Number(),
			),
			option.WithPermission(permission.StatusSeniorAdmin),
		),

		action.NewCommand(
			"prompt",
			h.ShowPrompt,
			i18n.Cmd.Chat.ShowPrompt.Desc,
			CategoryChat,
			option.WithAliases("промпт"),
			option.WithPredicates(predicate.SensitiveCommand()),
		),

		action.NewCommand(
			"setprompt",
			h.SetPrompt,
			i18n.Cmd.Chat.SetPrompt.Desc,
			CategoryChat,
			option.WithAliases("промпт"),
			option.WithRules(rule.Text()),
			option.WithPermission(permission.StatusSeniorAdmin),
		),

		action.NewCommand(
			"weekstart",
			h.ShowWeekStart,
			i18n.Cmd.Chat.ShowWeekStart.Desc,
			CategoryChat,
			option.WithAliases("начало недели"),
		),

		action.NewCommand(
			"setweekstart",
			h.SetWeekStart,
			i18n.Cmd.Chat.SetWeekStart.Desc,
			CategoryChat,
			option.WithAliases("начало недели"),
			option.WithRules(
				rule.DateTimeOrDuration(),
			),
			option.WithPermission(permission.StatusSeniorAdmin),
		),

		action.NewCommand(
			"prefix",
			h.ShowPrefix,
			i18n.Cmd.Chat.ShowPrefix.Desc,
			CategoryChat,
			option.WithAliases("префикс"),
		),

		action.NewCommand(
			"setprefix",
			h.SetPrefix,
			i18n.Cmd.Chat.SetPrefix.Desc,
			CategoryChat,
			option.WithAliases("префикс"),
			option.WithRules(
				rule.Text(),
			),
			option.WithPermission(permission.StatusSeniorAdmin),
		),

		action.NewCommand(
			"prefixonly",
			h.ShowPrefixlessStatus,
			i18n.Cmd.Chat.ShowPrefixless.Desc,
			CategoryChat,
		),

		action.NewCommand(
			"prefixonly",
			h.DisablePrefixOnly,
			i18n.Cmd.Chat.EnablePrefixless.Desc,
			CategoryChat,
			option.WithAliases("без префикса"),
		),

		action.NewCommand(
			"prefixfree",
			h.EnablePrefixOnly,
			i18n.Cmd.Chat.DisablePrefixless.Desc,
			CategoryChat,
			option.WithAliases("только с префиксом"),
		),

		action.NewCommand(
			"emojis",
			h.EnableEmojis,
			i18n.Cmd.Chat.EnableEmojis.Desc,
			CategoryChat,
			option.WithAliases("+эмодзи"),
		),

		action.NewCommand(
			"noemojis",
			h.DisableEmojis,
			i18n.Cmd.Chat.DisableEmojis.Desc,
			CategoryChat,
			option.WithAliases("-эмодзи"),
		),
		action.NewCommand("fakeleave", h.FakeLeave, i18n.Cmd.Chat.FakeLeave.Desc, CategoryChat, option.WithAliases("фейклив")),
	}
}

func (h *Handler) ShowNewbieThreshold(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.ShowNewbieThreshold.Success, i18n.CmdChatShowNewbieThresholdSuccessData{
		Days: ch.NewbieThresholdDays,
	}))
	return err
}
func (h *Handler) SetNewbieThreshold(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	days, ok := cctx.MustArgs(c).Number()
	if !ok {
		return nil
	}
	if days < 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.Chat.SetNewbieThreshold.Error.InvalidNumber, nil))
		return err
	}
	ch := cctx.MustChat(c)

	if err := h.service.SetNewbieThreshold(c, ch.ID, int32(days)); err != nil {
		return fmt.Errorf("set newbie threshold: %w", err)
	}

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.SetNewbieThreshold.Success, i18n.CmdChatSetNewbieThresholdSuccessData{
		Days: days,
	}))
	return err
}

func (h *Handler) SetPrompt(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}
	if err := h.service.SetChatPrompt(c, ch.ID, text); err != nil {
		return err
	}
	loc := cctx.MustLocalizer(c)
	_, err := c.Reply(loc.T(i18n.Cmd.Chat.SetPrompt.Success, nil))
	return err
}

func (h *Handler) ShowPrompt(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.ShowPrompt.Success, i18n.CmdChatShowPromptSuccessData{
		Prompt: ch.AISystemPrompt,
	}))
	return err
}

func (h *Handler) ShowWeekStart(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.Chat.ShowWeekStart.Success,
			i18n.CmdChatShowWeekStartSuccessData{
				Weekday: weekdayName(loc, ch.WeekStartDay),
				Time:    microsecondsToTime(ch.WeekStartTime),
			},
		),
	)

	return err
}

func microsecondsToTime(us int64) string {
	d := time.Duration(us) * time.Microsecond

	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour

	m := int(d / time.Minute)

	return fmt.Sprintf("%02d:%02d", h, m)
}

func weekdayName(loc *i18n.Localizer, day int16) string {
	switch day {
	case 1:
		return loc.T(i18n.Common.Weekday.Monday, nil)
	case 2:
		return loc.T(i18n.Common.Weekday.Tuesday, nil)
	case 3:
		return loc.T(i18n.Common.Weekday.Wednesday, nil)
	case 4:
		return loc.T(i18n.Common.Weekday.Thursday, nil)
	case 5:
		return loc.T(i18n.Common.Weekday.Friday, nil)
	case 6:
		return loc.T(i18n.Common.Weekday.Saturday, nil)
	case 7:
		return loc.T(i18n.Common.Weekday.Sunday, nil)
	default:
		return ""
	}
}

func (h *Handler) SetWeekStart(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	date, ok := cctx.MustArgs(c).Until()
	if !ok {
		return nil
	}

	weekday := date.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}

	if err := h.service.SetWeekStartDay(c, ch.ID, int(weekday)); err != nil {
		return fmt.Errorf("set week start day: %w", err)
	}

	newTime := fmt.Sprintf("%02d:%02d", date.Hour(), date.Minute())

	if err := h.service.SetWeekStartTime(c, ch.ID, newTime); err != nil {
		return fmt.Errorf("set week start time: %w", err)
	}

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.Chat.SetWeekStart.Success,
			i18n.CmdChatSetWeekStartSuccessData{
				Weekday: weekdayName(loc, int16(weekday)),
				Time:    tghtml.DateTime(date, "t", newTime),
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) ShowPrefix(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.Chat.ShowPrefix.Success.Custom,
			i18n.CmdChatShowPrefixSuccessCustomData{
				Prefix: ch.CommandPrefix,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) SetPrefix(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	prefix, ok := cctx.MustArgs(c).Text()
	if !ok || prefix == "" {
		_, err := c.Reply(loc.T(i18n.Cmd.Chat.SetPrefix.Error.Empty, nil))
		return err
	}

	if err := h.service.SetCommandPrefix(c, ch.ID, prefix); err != nil {
		return fmt.Errorf("set command prefix: %w", err)
	}

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.Chat.SetPrefix.Success.Custom,
			i18n.CmdChatSetPrefixSuccessCustomData{
				Prefix: prefix,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) DisablePrefixOnly(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if err := h.service.SetAllowPrefixless(c, ch.ID, true); err != nil {
		return fmt.Errorf("set allow prefixless: %w", err)
	}

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.EnablePrefixless.Success, nil))
	return err
}

func (h *Handler) EnablePrefixOnly(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if err := h.service.SetAllowPrefixless(c, ch.ID, false); err != nil {
		return fmt.Errorf("set allow prefixless: %w", err)
	}

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.DisablePrefixless.Success, nil))
	return err
}

func (h *Handler) ShowPrefixlessStatus(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	key := i18n.Cmd.Chat.ShowPrefixless.Disabled
	if ch.AllowPrefixless {
		key = i18n.Cmd.Chat.ShowPrefixless.Enabled
	}

	_, err := c.Reply(loc.T(key, nil))
	return err
}

func (h *Handler) DisableEmojis(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if err := h.service.SetEmojisEnabled(c, ch.ID, false); err != nil {
		return fmt.Errorf("set emojis enabled: %w", err)
	}

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.DisableEmojis.Success, nil))
	return err
}

func (h *Handler) EnableEmojis(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if err := h.service.SetEmojisEnabled(c, ch.ID, true); err != nil {
		return fmt.Errorf("set emojis enabled: %w", err)
	}

	_, err := c.Reply(loc.T(i18n.Cmd.Chat.EnableEmojis.Success, nil))
	return err
}

func (h *Handler) FakeLeave(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}
	cm := cctx.MustChatMember(c)
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	chatID, _ := c.Chat()
	_ = c.Bot.DeleteMessage(c, chatID, msg.MessageID)
	text := loc.T(i18n.User.Left, i18n.UserLeftData{
		User: tghtml.MemberMention(loc, ch, cm),
	}, i18n.WithGender(cm.Gender()))

	if _, err := c.Bot.SendMessage(c, chatID, text, botapi.DisableWebPagePreview(), botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
		return err
	}
	return nil
}

package rest

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

func (h *Handler) ShowRest(c *botapi.Context) error {
	cm, err := cctx.ChatMember(c.Context)
	if err != nil {
		return fmt.Errorf("show rest: %w", err)
	}

	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("show rest: %w", err)
	}

	cmLink := tghtml.MemberLink(h.translator, ch, cm)

	if !cm.IsResting(time.Now()) {
		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.Rest.NoRest, i18n.CmdRestNoRestArgs(cmLink)),
			botapi.WithParseMode(botapi.ParseModeHTML),
			botapi.DisableWebPagePreview(),
		)

		return err
	}

	_, err = c.Reply(h.translator.TData(
		ch.Lang, i18n.Cmd.Rest.Info,
		i18n.CmdRestInfoArgs(
			tghtml.DateTime(cm.RestUntil, "wdt", cm.RestUntil.Format("02.01.2006")),
			cmLink,
		),
	),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

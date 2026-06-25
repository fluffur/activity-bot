package stats

import (
	"activity-bot/internal/norm"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
)

type Presenter struct {
	translator *i18n.Translator
}

func NewPresenter(t *i18n.Translator) *Presenter {
	return &Presenter{translator: t}
}

func (p *Presenter) RenderStats(ch chat.Chat, data CalculatedStats) string {
	var b strings.Builder

	title := p.translator.TData(ch.Lang, i18n.Cmd.Stats.Title, i18n.CmdStatsTitleArgs(
		tghtml.DateTime(data.FromDate, "wdt", data.FromDate.Format("02.01.2006")),
		tghtml.DateTime(data.ToDate, "wdt", data.ToDate.Format("02.01.2006")),
	))
	b.WriteString(title)
	b.WriteString("\n\n")

	if !data.HasNorms {
		b.WriteString("<blockquote expandable>")
		for i, u := range data.SimpleResults {
			b.WriteString(fmt.Sprintf(
				"%d. %s — %d",
				i+1,
				tghtml.MemberLink(p.translator, ch, u.Member),
				u.Messages,
			))
			if i != len(data.SimpleResults)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("</blockquote>\n\n")
	} else {
		for _, r := range data.NormResults {
			b.WriteString(p.translator.TData(ch.Lang, i18n.Cmd.Stats.NormTitle, i18n.CmdStatsNormTitleArgs(
				tghtml.Bold(UcFirst(norm.LocalisedNormName(p.translator, ch.Lang, r.NormName))),
				tghtml.Code(fmt.Sprintf("%d", r.Required)),
			)))
			b.WriteString("\n")

			if len(r.Failed) > 0 {
				var failed strings.Builder
				for i, u := range r.Failed {
					failed.WriteString(p.translator.TData(ch.Lang, i18n.Cmd.Stats.UserFailed, i18n.CmdStatsUserFailedArgs(
						i+1,
						u.Messages,
						r.Required,
						tghtml.MemberLink(p.translator, ch, u.Member),
					)) + "\n")
				}
				b.WriteString(p.translator.T(ch.Lang, i18n.Cmd.Stats.Failed) + "\n")
				b.WriteString(tghtml.ExpandableBlockquote(failed.String()) + "\n")
			}

			if len(r.Passed) > 0 {
				var passed strings.Builder
				for i, u := range r.Passed {
					passed.WriteString(p.translator.TData(ch.Lang, i18n.Cmd.Stats.UserPassed, i18n.CmdStatsUserPassedArgs(
						i+1,
						u.Messages,
						tghtml.MemberLink(p.translator, ch, u.Member),
					)) + "\n")
				}
				b.WriteString(p.translator.T(ch.Lang, i18n.Cmd.Stats.Passed) + "\n")
				b.WriteString(tghtml.ExpandableBlockquote(passed.String()) + "\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(p.translator.TData(
		ch.Lang,
		i18n.Cmd.Stats.TotalMessages,
		i18n.CmdStatsTotalMessagesArgs(data.TotalMessages),
	))

	return b.String()
}

func UcFirst(s string) string {
	if s == "" {
		return ""
	}

	r, size := utf8.DecodeRuneInString(s)

	rUpper := unicode.ToUpper(r)

	var b strings.Builder
	b.Grow(len(s))
	b.WriteRune(rUpper)
	b.WriteString(s[size:])

	return b.String()
}

package stats

import (
	"activity-bot/internal/norm"
	"fmt"
	"strings"
	"time"
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
			b.WriteString("\n\n")

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

func (p *Presenter) RenderProfile(
	ch chat.Chat,
	profile ProfileStats,
) string {
	var b strings.Builder

	now := time.Now()

	status := profile.ChatMember.Status

	b.WriteString(
		p.translator.TData(
			ch.Lang,
			i18n.Cmd.Profile.Title,
			i18n.CmdProfileTitleArgs(
				tghtml.MemberLink(
					p.translator,
					ch,
					profile.ChatMember,
				),
			),
		),
	)

	b.WriteString(" · ")

	b.WriteString(
		fmt.Sprintf("%s %s", status.Emoji(), p.translator.T(ch.Lang, status.TranslationKey())),
	)

	b.WriteString("\n\n")

	if profile.ChatMember.LeftAt.IsZero() {
		days := int(now.Sub(profile.ChatMember.JoinedAt).Hours() / 24)

		b.WriteString(
			p.translator.TData(
				ch.Lang,
				i18n.Cmd.Profile.MemberSince,
				i18n.CmdProfileMemberSinceArgs(
					tghtml.DateTime(
						profile.ChatMember.JoinedAt,
						"wdt",
						profile.ChatMember.JoinedAt.Format("02.01.2006"),
					),
					tghtml.DateTime(
						profile.ChatMember.JoinedAt,
						"r",
						fmt.Sprintf("%d дн.", days),
					),
				),
			),
		)
		b.WriteString("\n")
	} else {
		days := int(profile.ChatMember.LeftAt.Sub(profile.ChatMember.JoinedAt).Hours() / 24)

		b.WriteString(
			p.translator.TData(
				ch.Lang,
				i18n.Cmd.Profile.MemberPeriod,
				i18n.CmdProfileMemberPeriodArgs(
					tghtml.DateTime(
						profile.ChatMember.JoinedAt,
						"wdt",
						profile.ChatMember.JoinedAt.Format("02.01.2006"),
					),
					tghtml.DateTime(
						profile.ChatMember.LeftAt,
						"wdt",
						profile.ChatMember.LeftAt.Format("02.01.2006"),
					),
					fmt.Sprintf("%d дн.", days),
				),
			),
		)
		b.WriteString("\n")
	}
	restActive := profile.ChatMember.RestUntil.After(now)

	if restActive {
		b.WriteString(
			p.translator.TData(
				ch.Lang,
				i18n.Cmd.Profile.RestUntil,
				i18n.CmdProfileRestUntilArgs(
					tghtml.DateTime(
						profile.ChatMember.RestUntil,
						"wdt",
						profile.ChatMember.RestUntil.Format("02.01.2006"),
					),
				),
			),
		)
		b.WriteString("\n")

		b.WriteString(
			p.translator.T(
				ch.Lang,
				i18n.Cmd.Profile.RestExempt,
			),
		)
		b.WriteString("\n")
	}

	var content strings.Builder

	content.WriteString(
		p.translator.TData(
			ch.Lang,
			i18n.Cmd.Profile.Activity,
			i18n.CmdProfileActivityArgs(
				profile.DayCount,
				profile.DayRollingCount,
				profile.WeekCount,
				profile.WeekRollingCount,
				profile.MonthCount,
				profile.MonthRollingCount,
				profile.AllTimeCount,
			),
		),
	)

	b.WriteString("\n")
	b.WriteString(
		tghtml.ExpandableBlockquote(
			content.String(),
		),
	)

	if !restActive && len(profile.Norms) > 0 {
		b.WriteString("\n\n")

		for _, n := range profile.Norms {
			normName := UcFirst(
				norm.LocalisedNormName(
					p.translator,
					ch.Lang,
					n.Name,
				),
			)

			key, arg := i18n.Cmd.Profile.NormFailed,
				i18n.CmdProfileNormFailedArgs(
					n.Current,
					normName,
					n.Required,
				)

			if n.Passed {
				key, arg = i18n.Cmd.Profile.NormPassed,
					i18n.CmdProfileNormPassedArgs(
						normName,
						n.Required,
					)
			}

			b.WriteString(
				p.translator.TData(
					ch.Lang,
					key,
					arg,
				),
			)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

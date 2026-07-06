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

func RenderStats(loc *i18n.Localizer, ch chat.Chat, data CalculatedStats) string {
	var b strings.Builder

	b.WriteString(loc.T(
		i18n.Cmd.Stats.Title,
		i18n.CmdStatsTitleData{
			From: tghtml.DefaultDateTime(data.FromDate),
			To:   tghtml.DefaultDateTime(data.ToDate),
		},
	))
	b.WriteString("\n\n")

	if !data.HasNorms {
		b.WriteString("<blockquote expandable>")

		for i, u := range data.SimpleResults {
			_, _ = fmt.Fprintf(
				&b,
				"%d. %s — %d",
				i+1,
				tghtml.MemberLink(loc, ch, u.Member),
				u.Messages,
			)

			if i != len(data.SimpleResults)-1 {
				b.WriteByte('\n')
			}
		}

		b.WriteString("</blockquote>\n\n")
	} else {
		for _, r := range data.NormResults {
			b.WriteString(loc.T(
				i18n.Cmd.Stats.NormTitle,
				i18n.CmdStatsNormTitleData{
					Name:     tghtml.Bold(UcFirst(norm.LocalisedNormName(loc, r.NormName))),
					Required: tghtml.Code(fmt.Sprintf("%d", r.Required)),
				},
			))
			b.WriteString("\n\n")

			if len(r.Failed) > 0 {
				var failed strings.Builder

				for i, u := range r.Failed {
					failed.WriteString(loc.T(
						i18n.Cmd.Stats.UserFailed,
						i18n.CmdStatsUserFailedData{
							List:     i + 1,
							User:     tghtml.MemberLink(loc, ch, u.Member),
							Messages: u.Messages,
							Required: r.Required,
						},
					))
					failed.WriteByte('\n')
				}

				b.WriteString(loc.T(i18n.Cmd.Stats.Failed, nil))
				b.WriteByte('\n')
				b.WriteString(tghtml.ExpandableBlockquote(failed.String()))
				b.WriteByte('\n')
			}

			if len(r.Passed) > 0 {
				var passed strings.Builder

				for i, u := range r.Passed {
					passed.WriteString(loc.T(
						i18n.Cmd.Stats.UserPassed,
						i18n.CmdStatsUserPassedData{
							List:     i + 1,
							User:     tghtml.MemberLink(loc, ch, u.Member),
							Messages: u.Messages,
						},
					))
					passed.WriteByte('\n')
				}

				b.WriteString(loc.T(i18n.Cmd.Stats.Passed, nil))
				b.WriteByte('\n')
				b.WriteString(tghtml.ExpandableBlockquote(passed.String()))
				b.WriteByte('\n')
			}

			b.WriteByte('\n')
		}
	}

	b.WriteString(loc.T(
		i18n.Cmd.Stats.TotalMessages,
		i18n.CmdStatsTotalMessagesData{
			Total: data.TotalMessages,
		},
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

func RenderProfile(
	loc *i18n.Localizer,
	ch chat.Chat,
	profile ProfileStats,
) string {
	var b strings.Builder

	now := time.Now()
	status := profile.ChatMember.Status

	b.WriteString(loc.T(
		i18n.Cmd.Profile.Title,
		i18n.CmdProfileTitleData{
			User: tghtml.MemberLink(loc, ch, profile.ChatMember),
		},
	))

	b.WriteString(" · ")
	b.WriteString(fmt.Sprintf("%s %s", status.Emoji(), loc.T(status.TranslationKey(), nil)))
	b.WriteString("\n\n")

	if profile.ChatMember.LeftAt.IsZero() {
		b.WriteString(loc.T(
			i18n.Cmd.Profile.MemberSince,
			i18n.CmdProfileMemberSinceData{
				Date: tghtml.DefaultDateTime(profile.ChatMember.JoinedAt),
				Days: tghtml.RelativeDateTime(profile.ChatMember.JoinedAt, now),
			},
		))
		b.WriteString("\n")
	} else {
		days := int(profile.ChatMember.LeftAt.Sub(profile.ChatMember.JoinedAt).Hours() / 24)

		b.WriteString(loc.T(
			i18n.Cmd.Profile.MemberPeriod,
			i18n.CmdProfileMemberPeriodData{
				From: tghtml.DefaultDateTime(profile.ChatMember.JoinedAt),
				To:   tghtml.DefaultDateTime(profile.ChatMember.LeftAt),
				Days: fmt.Sprintf("%d дн.", days),
			},
		))
		b.WriteString("\n")
	}

	restActive := profile.ChatMember.RestUntil.After(now)

	if restActive {
		b.WriteString(loc.T(
			i18n.Cmd.Profile.RestUntil,
			i18n.CmdProfileRestUntilData{
				Date: tghtml.DefaultDateTime(profile.ChatMember.RestUntil),
			},
		))
		b.WriteString("\n")

		b.WriteString(loc.T(i18n.Cmd.Profile.RestExempt, nil))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(
		tghtml.ExpandableBlockquote(
			loc.T(
				i18n.Cmd.Profile.Activity,
				i18n.CmdProfileActivityData{
					Day:          tghtml.Number(profile.DayCount),
					Week:         tghtml.Number(profile.WeekCount),
					Month:        tghtml.Number(profile.MonthCount),
					DayRolling:   tghtml.Number(profile.DayRollingCount),
					WeekRolling:  tghtml.Number(profile.WeekRollingCount),
					MonthRolling: tghtml.Number(profile.MonthRollingCount),
					Total:        tghtml.Number(profile.AllTimeCount),
				},
			),
		),
	)

	if !restActive && len(profile.Norms) > 0 {
		b.WriteString("\n\n")

		for _, n := range profile.Norms {
			normName := UcFirst(norm.LocalisedNormName(loc, n.Name))

			if n.Passed {
				b.WriteString(loc.T(
					i18n.Cmd.Profile.NormPassed,
					i18n.CmdProfileNormPassedData{
						Name:     normName,
						Required: n.Required,
					},
				))
			} else {
				b.WriteString(loc.T(
					i18n.Cmd.Profile.NormFailed,
					i18n.CmdProfileNormFailedData{
						Name:     normName,
						Current:  n.Current,
						Required: n.Required,
					},
				))
			}

			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

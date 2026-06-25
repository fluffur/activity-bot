package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/norm"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gotd/botapi"
)

type UserNormStat struct {
	Norm     string
	Required int32
	Actual   int64
	Passed   bool
}

type UserStats struct {
	ChatMember chatmember.ChatMember
	Norms      []UserNormStat
}

type UserResult struct {
	Member   chatmember.ChatMember
	Messages int64
}

type NormResult struct {
	NormName string
	Required int32

	Passed []UserResult
	Failed []UserResult
}

func (h *Handler) Chat(c *botapi.Context) error {
	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("chat no args")
	}

	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("stats chat: %w", err)
	}

	var (
		fromDate time.Time
		toDate   time.Time
	)

	now := time.Now()
	switch {
	case len(args.Durations) == 1:
		fromDate = now.Add(-args.Durations[0])
		toDate = now

	case len(args.DateTimes) == 2:
		fromDate = args.DateTimes[0]
		toDate = args.DateTimes[1]

	default:
		from, to, err := currentChatWeekRange(
			ch.WeekStartDay,
			ch.WeekStartTimeMicros,
		)
		if err != nil {
			return fmt.Errorf("stats chat: %w", err)
		}

		fromDate = from
		toDate = to
	}

	chatMembers, err := h.chatMemberRepository.List(c.Context, chatmember.Filter{
		ChatID: ch.ID,
		IsBot: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
		Left: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("stats members list: %w", err)
	}

	norms, err := h.normRepository.ListWithMembers(c.Context, ch.ID)
	if err != nil {
		return fmt.Errorf("stats norms: %w", err)
	}

	chatStats, err := h.statsRepository.ChatStats(c.Context, ch.ID, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("chat stats: %w", err)
	}

	statsByUserID := make(map[int64]int64)
	var totalMessages int64

	for _, stat := range chatStats {
		statsByUserID[stat.ChatMember.User.ID] = stat.MessagesCount
		totalMessages += stat.MessagesCount
	}
	userNorms := make(map[int64][]norm.Norm)
	var commonNorms []norm.Norm

	title := h.translator.TData(
		ch.Lang,
		i18n.Cmd.Stats.Title,
		i18n.CmdStatsTitleArgs(
			tghtml.DateTime(
				fromDate,
				"wdt",
				fromDate.Format("02.01.2006"),
			),
			tghtml.DateTime(
				toDate,
				"wdt",
				toDate.Format("02.01.2006"),
			),
		),
	)

	if len(norms) == 0 {
		var results []UserResult

		for _, member := range chatMembers {
			if member.IsResting(now) || member.IsNewbie(now, ch.NewbieThresholdDays) {
				continue
			}

			results = append(results, UserResult{
				Member:   member,
				Messages: statsByUserID[member.User.ID],
			})
		}

		slices.SortFunc(results, func(a, b UserResult) int {
			switch {
			case a.Messages > b.Messages:
				return -1
			case a.Messages < b.Messages:
				return 1
			default:
				return 0
			}
		})

		var b strings.Builder

		b.WriteString(title)
		b.WriteString("\n\n")
		b.WriteString("<blockquote expandable>")
		for i, u := range results {
			b.WriteString(
				fmt.Sprintf(
					"%d. %s — %d",
					i+1,
					tghtml.MemberLink(h.translator, ch, u.Member),
					u.Messages,
				),
			)
			if i != len(results)-1 {
				b.WriteString("\n")
			}
		}

		b.WriteString("</blockquote>\n")

		b.WriteString("\n")
		b.WriteString(
			h.translator.TData(ch.Lang, i18n.Cmd.Stats.TotalMessages, i18n.CmdStatsTotalMessagesArgs(totalMessages)),
		)

		_, err := c.Reply(
			b.String(),
			botapi.WithParseMode(botapi.ParseModeHTML),
			botapi.DisableWebPagePreview(),
		)

		return err
	}

	for _, n := range norms {
		if len(n.UserIDs) == 0 {
			commonNorms = append(commonNorms, n)
			continue
		}

		for _, userID := range n.UserIDs {
			userNorms[userID] = append(userNorms[userID], n)
		}
	}

	normResults := make(map[int64]*NormResult)

	for _, n := range norms {
		normResults[n.ID] = &NormResult{
			NormName: n.Name,
			Required: n.Value,
		}
	}

	for _, member := range chatMembers {
		if member.IsResting(now) || member.IsNewbie(now, ch.NewbieThresholdDays) {
			continue
		}

		userID := member.User.ID
		messages := statsByUserID[userID]

		memberNorms := userNorms[userID]
		if len(memberNorms) == 0 {
			memberNorms = commonNorms
		}

		for _, n := range memberNorms {

			r := normResults[n.ID]
			userResult := UserResult{
				Member:   member,
				Messages: messages,
			}

			if messages >= int64(n.Value) {
				r.Passed = append(r.Passed, userResult)
			} else {
				r.Failed = append(r.Failed, userResult)
			}
		}
	}

	var b strings.Builder

	b.WriteString(title)
	b.WriteString("\n\n")

	for _, n := range norms {
		r := normResults[n.ID]

		b.WriteString(
			h.translator.TData(
				ch.Lang,
				i18n.Cmd.Stats.NormTitle,
				i18n.CmdStatsNormTitleArgs(
					norm.LocalisedNormName(
						h.translator,
						ch.Lang,
						r.NormName,
					),
					tghtml.Code(fmt.Sprintf("%d", r.Required)),
				),
			),
		)
		b.WriteString("\n")

		if len(r.Failed) > 0 {
			var failed strings.Builder

			for i, u := range r.Failed {
				failed.WriteString(
					h.translator.TData(
						ch.Lang,
						i18n.Cmd.Stats.UserFailed,
						i18n.CmdStatsUserFailedArgs(
							i+1,
							u.Messages,
							r.Required,
							tghtml.MemberLink(h.translator, ch, u.Member),
						),
					),
				)
				failed.WriteString("\n")
			}

			b.WriteString(
				h.translator.T(
					ch.Lang,
					i18n.Cmd.Stats.Failed,
				),
			)
			b.WriteString("\n")

			b.WriteString(
				tghtml.ExpandableBlockquote(
					failed.String(),
				),
			)
			b.WriteString("\n")
		}

		if len(r.Passed) > 0 {
			var passed strings.Builder

			for i, u := range r.Passed {
				passed.WriteString(
					h.translator.TData(
						ch.Lang,
						i18n.Cmd.Stats.UserPassed,
						i18n.CmdStatsUserPassedArgs(
							i+1,
							u.Messages,
							tghtml.MemberLink(h.translator, ch, u.Member),
						),
					),
				)
				passed.WriteString("\n")
			}

			b.WriteString(
				h.translator.T(
					ch.Lang,
					i18n.Cmd.Stats.Passed,
				),
			)
			b.WriteString("\n")

			b.WriteString(
				tghtml.ExpandableBlockquote(
					passed.String(),
				),
			)
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	b.WriteString(
		h.translator.TData(ch.Lang, i18n.Cmd.Stats.TotalMessages, i18n.CmdStatsTotalMessagesArgs(totalMessages)),
	)

	if _, err := c.Reply(b.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview()); err != nil {
		return err
	}

	return nil
}

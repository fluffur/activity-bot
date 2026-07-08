package rest

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/botapi"
)

func (h *Handler) ShowRest(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	cm := cctx.MustChatMember(c)
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if len(args.Users) != 0 {
		cm = args.Users[0]
	}

	cmMention := tghtml.MemberMention(loc, ch, cm)

	if !cm.IsResting(time.Now()) {
		_, err := c.Reply(loc.T(i18n.Cmd.Rest.NoRest, i18n.CmdRestNoRestData{User: cmMention}),
			botapi.WithParseMode(botapi.ParseModeHTML),
			botapi.DisableWebPagePreview(),
		)

		return err
	}

	restUntil := tghtml.DefaultDateTime(cm.RestUntil)

	var text string

	if cm.RestReason == "" {
		text = loc.T(i18n.Cmd.Rest.Info, i18n.CmdRestInfoData{
			User: cmMention,
			Date: restUntil,
		})
	} else {
		text = loc.T(i18n.Cmd.Rest.InfoReason, i18n.CmdRestInfoReasonData{
			User:   cmMention,
			Reason: cm.RestReason,
			Date:   restUntil,
		})
	}

	_, err := c.Reply(text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func (h *Handler) SetRest(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	ch := cctx.MustChat(c)
	self := cctx.MustChatMember(c)
	p := cctx.MustPermission(c)
	loc := cctx.MustLocalizer(c)

	cm := self
	if len(args.Users) > 0 {
		cm = args.Users[0]
	}

	until, err := ParseRestUntil(args)
	if err != nil {
		return fmt.Errorf("set rest: %w", err)
	}

	msg := c.Message()
	if msg == nil {
		return nil
	}

	var reason string

	if len(args.Texts) != 0 {
		reason = args.Texts[0]
	}

	if !self.Permitted(p) {
		return h.createRestRequest(c, ch, cm, until, reason)
	}

	if err := h.service.SetMemberRestWithHistory(
		c,
		ch.ID,
		cm.User.ID,
		int64(msg.MessageID),
		until,
		reason,
	); err != nil {
		return fmt.Errorf("set rest: %w", err)
	}

	mention := tghtml.MemberMention(loc, ch, cm)
	date := tghtml.DefaultDateTime(until)

	var text string

	if reason == "" {
		text = loc.T(
			i18n.Cmd.Rest.Set,
			i18n.CmdRestSetData{
				User: mention,
				Date: date,
			},
		)
	} else {
		text = loc.T(
			i18n.Cmd.Rest.SetReason,
			i18n.CmdRestSetReasonData{
				User:   mention,
				Date:   date,
				Reason: reason,
			},
		)
	}

	_, err = c.Reply(
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func ParseRestUntil(args cctx.ParsedArgs) (time.Time, error) {
	switch {
	case len(args.DateTimes) == 1:
		return args.DateTimes[0], nil

	case len(args.Durations) == 1:
		return time.Now().Add(args.Durations[0]), nil

	default:
		return time.Time{}, fmt.Errorf("rest date not found")
	}
}

func (h *Handler) EndRest(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	loc := cctx.MustLocalizer(c)

	if len(args.Users) > 0 {
		cm = args.Users[0]
	}

	if !cm.IsResting(time.Now()) {
		_, err := c.Reply(
			loc.T(
				i18n.Cmd.EndRest.NotInRest,
				i18n.CmdEndRestNotInRestData{
					User: tghtml.MemberMention(loc, ch, cm),
				},
			),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	if err := h.service.EndMemberRest(
		c,
		ch.ID,
		cm.User.ID,
	); err != nil {
		return fmt.Errorf("end rest: %w", err)
	}

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.EndRest.Ended,
			i18n.CmdEndRestEndedData{
				User: tghtml.MemberMention(loc, ch, cm),
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) AllUserRests(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)

	if len(args.Users) > 0 {
		cm = args.Users[0]
	}

	requests, err := h.service.GetRequests(
		c,
		ch.ID,
		cm.User.ID,
	)
	if err != nil {
		return fmt.Errorf("all rests: %w", err)
	}

	if len(requests) == 0 {
		_, err = c.Reply(
			loc.T(
				i18n.Cmd.Rests.HistoryEmpty,
				i18n.CmdRestsHistoryEmptyData{
					User: tghtml.MemberMention(loc, ch, cm),
				},
			),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	var text strings.Builder

	text.WriteString(
		loc.T(
			i18n.Cmd.Rests.HistoryTitle,
			i18n.CmdRestsHistoryTitleData{
				User: tghtml.MemberMention(loc, ch, cm),
			},
		),
	)
	text.WriteString("\n\n<blockquote expandable>")

	for i, req := range requests {
		if i > 0 {
			text.WriteString("\n\n")
		}

		if req.Reason == "" {
			text.WriteString(
				loc.T(
					i18n.Cmd.Rests.HistoryItem,
					i18n.CmdRestsHistoryItemData{
						Index:    i + 1,
						From:     tghtml.DefaultDateTime(req.RequestedAt),
						To:       tghtml.DefaultDateTime(req.RestUntil),
						Duration: tghtml.RelativeDateTime(req.RequestedAt, req.RestUntil),
					},
				),
			)
		} else {
			text.WriteString(
				loc.T(
					i18n.Cmd.Rests.HistoryItemReason,
					i18n.CmdRestsHistoryItemReasonData{
						Index:    i + 1,
						From:     tghtml.DefaultDateTime(req.RequestedAt),
						To:       tghtml.DefaultDateTime(req.RestUntil),
						Duration: tghtml.RelativeDateTime(req.RequestedAt, req.RestUntil),
						Reason:   req.Reason,
					},
				),
			)
		}
	}

	text.WriteString("</blockquote>")

	text.WriteString(
		loc.T(
			i18n.Cmd.Rests.HistoryTotal,
			i18n.CmdRestsHistoryTotalData{
				Count: len(requests),
			},
		),
	)

	_, err = c.Reply(
		text.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) ApproveRestRequest(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	data := cq.Data

	userID, err := parseRequestCallbackData(data)
	if err != nil {
		return fmt.Errorf("approve rest request: %w", err)
	}

	request, err := h.service.GetRestRequest(
		c,
		ch.ID,
		userID,
		int64(cq.Message.MessageID),
	)
	if err != nil {
		_ = c.AnswerCallback(
			botapi.WithCallbackText(
				loc.T(i18n.Cmd.RestRequest.NoRequest, nil),
			),
		)

		return fmt.Errorf("get rest request: %w", err)
	}

	if err := h.service.ApproveRestRequest(
		c,
		ch.ID,
		userID,
		int64(cq.Message.MessageID),
		request.RestUntil,
	); err != nil {
		_ = c.AnswerCallback()
		return fmt.Errorf("approve rest request: %w", err)
	}

	member, err := h.chatMemberService.Get(
		c,
		ch.ID,
		userID,
	)
	if err != nil {
		_ = c.AnswerCallback()
		return fmt.Errorf("approve rest request chat member: %w", err)
	}

	text := loc.T(
		i18n.Cmd.RestRequest.Approved,
		i18n.CmdRestRequestApprovedData{
			User: tghtml.MemberMention(loc, ch, member),
			Date: tghtml.DefaultDateTime(request.RestUntil),
		},
	)

	chatID, ok := c.Chat()
	if !ok {
		_ = c.AnswerCallback()
		return fmt.Errorf("approve no chat")
	}

	_, err = c.Bot.EditMessageText(
		c,
		chatID,
		cq.Message.MessageID,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)
	if err != nil {
		return fmt.Errorf("edit approved request: %w", err)
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(
			loc.T(i18n.Cmd.RestRequest.Approved, nil),
		),
	)
}

func (h *Handler) RejectRestRequest(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	p := cctx.MustPermission(c)
	loc := cctx.MustLocalizer(c)

	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	data := cq.Data

	userID, err := parseRequestCallbackData(data)
	if err != nil {
		return fmt.Errorf("reject rest request: %w", err)
	}

	if !cm.Permitted(p) && cm.User.ID != userID {
		return c.AnswerCallback(
			botapi.WithCallbackText(
				loc.T(
					i18n.System.NoPermission,
					i18n.SystemNoPermissionData{
						Status: loc.T(p.TranslationKey(), nil),
					},
				),
			),
		)
	}

	if err := h.service.RejectRestRequest(
		c,
		ch.ID,
		userID,
		int64(cq.Message.MessageID),
	); err != nil {
		_ = c.AnswerCallback()
		return fmt.Errorf("approve rest request: %w", err)
	}

	text := loc.T(i18n.Cmd.RestRequest.Rejected, nil)

	chatID, ok := c.Chat()
	if !ok {
		_ = c.AnswerCallback()
		return nil
	}

	_, err = c.Bot.EditMessageText(
		c,
		chatID,
		cq.Message.MessageID,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)
	if err != nil {
		return fmt.Errorf("edit approved request: %w", err)
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(text),
	)
}

func (h *Handler) RemoveRestRequest(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	cm := cctx.MustChatMember(c)
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	var number int

	if len(args.Numbers) > 0 {
		number = int(args.Numbers[0])
	}

	if number == 0 {
		return nil
	}

	target := cm
	if len(args.Users) != 0 {
		target = args.Users[0]
	}

	requests, err := h.service.GetRequests(c, ch.ID, target.ID())
	if err != nil {
		return fmt.Errorf("remove rest request: get requests: %w", err)
	}

	if number > len(requests) {
		return nil
	}

	request := requests[number-1]

	if request.RestUntil.Equal(target.RestUntil) {
		if err := h.service.DeleteRestRequestAndEndRest(c, ch.ID, request.UserID, request.ID); err != nil {
			return fmt.Errorf("remove rest request: delete and end rest: %w", err)
		}

		_, err = c.Reply(
			loc.T(i18n.Cmd.Rests.Delete.DeletedActive, nil),
		)

		return err
	}

	if err := h.service.DeleteRestRequest(c, request.ID); err != nil {
		return fmt.Errorf("remove rest request: delete request: %w", err)
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.Rests.Delete.Deleted, nil),
	)

	return err
}

func parseRequestCallbackData(data string) (int64, error) {
	parts := strings.SplitN(data, ":", 3)

	if len(parts) != 3 {
		return 0, errors.New("invalid callback data")
	}

	return strconv.ParseInt(parts[2], 10, 64)
}

func (h *Handler) createRestRequest(
	c *botapi.Context,
	ch chat.Chat,
	cm chatmember.ChatMember,
	until time.Time,
	reason string,
) error {
	loc := cctx.MustLocalizer(c)

	mention := tghtml.MemberMention(loc, ch, cm)
	date := tghtml.DefaultDateTime(until)

	var text string

	if reason == "" {
		text = loc.T(
			i18n.Cmd.RestRequest.Text,
			i18n.CmdRestRequestTextData{
				User: mention,
				Date: date,
			},
		)
	} else {
		text = loc.T(
			i18n.Cmd.RestRequest.TextReason,
			i18n.CmdRestRequestTextReasonData{
				User:   mention,
				Date:   date,
				Reason: reason,
			},
		)
	}

	msg, err := c.Reply(
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						loc.T(i18n.Cmd.RestRequest.ApproveButton, nil),
						fmt.Sprintf("%s%d", callbackRestApprove, cm.User.ID),
					),
					botapi.InlineButtonData(
						loc.T(i18n.Cmd.RestRequest.RejectButton, nil),
						fmt.Sprintf("%s%d", callbackRestReject, cm.User.ID),
					),
				),
			),
		),
	)
	if err != nil {
		return fmt.Errorf("send rest request: %w", err)
	}

	if err := h.service.CreateRestRequest(
		c,
		ch.ID,
		cm.User.ID,
		int64(msg.MessageID),
		until,
	); err != nil {
		return fmt.Errorf("create rest request: %w", err)
	}

	return nil
}

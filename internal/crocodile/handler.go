package crocodile

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/utils/tghtml"
	"errors"

	"github.com/gotd/botapi"
)

const CategoryCrocodile command.Category = "crocodile"

const (
	callbackNextWord   = "crocodile_next"
	callbackShowWord   = "crocodile_show"
	callbackFinish     = "crocodile_finish"
	callbackBecomeHost = "crocodile_host"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"crocodile",
			h.Start,
			i18n.Cmd.Crocodile.Desc,
			CategoryCrocodile,
			option.WithAliases("крокодил", "крискадил"),
		),

		action.NewCommand(
			"stopcrocodile",
			h.Stop,
			i18n.Cmd.Crocodile.Stop.Desc,
			CategoryCrocodile,
			option.WithAliases("стоп крокодил"),
		),
		action.NewCallback(
			"crocodilenext",
			callbackNextWord,
			h.NextWord,
			CategoryCrocodile,
		),

		action.NewCallback(
			"crocodileshow",
			callbackShowWord,
			h.ShowWord,
			CategoryCrocodile,
		),
		action.NewCallback(
			"crocodilefinish",
			callbackFinish,
			h.Finish,
			CategoryCrocodile,
		),
		action.NewMessage(
			h.HandleMessage,
			CategoryCrocodile,
			option.WithPredicates(IsCrocodileGameActive(h.service)),
		),
		action.NewCallback(
			"crocodilehost",
			callbackBecomeHost,
			h.BecomeHost,
			CategoryCrocodile,
		),
	}
}

func IsCrocodileGameActive(service *Service) botapi.Predicate {
	return func(c *botapi.Context) bool {
		ch := cctx.MustChat(c)

		_, err := service.Get(c, ch.ID)

		return err == nil
	}
}

func (h *Handler) Start(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)

	if _, err := h.service.Start(c, ch.ID, cm.ID()); err != nil {
		_, err = c.Reply(loc.T(i18n.Cmd.Crocodile.Error.AlreadyStarted, nil))
		return err
	}

	keyboard := botapi.InlineKeyboard(
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.ShowWord, nil),
				callbackShowWord,
			),
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.Next, nil),
				callbackNextWord,
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.Finish, nil),
				callbackFinish,
			),
		),
	)

	_, err := c.Reply(
		loc.T(
			i18n.Cmd.Crocodile.Started,
			i18n.CmdCrocodileStartedData{
				Host: tghtml.MemberMention(
					loc,
					ch,
					cctx.MustChatMember(c),
				),
			},
		),
		botapi.WithReplyMarkup(keyboard),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)
	return err
}

func (h *Handler) ShowWord(c *botapi.Context) error {
	cb := c.Update.CallbackQuery
	if cb == nil {
		return nil
	}

	game, err := h.service.Get(c, cb.Message.Chat.ID)
	if err != nil {
		return err
	}

	loc := cctx.MustLocalizer(c)
	if game.HostID != cb.From.ID {
		return c.AnswerCallback(
			botapi.WithCallbackText(
				loc.T(i18n.Cmd.Crocodile.Error.NotHost, nil),
			),
		)
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(game.Word),
		botapi.WithCallbackAlert(),
	)
}

func (h *Handler) NextWord(c *botapi.Context) error {
	cb := c.Update.CallbackQuery
	if cb == nil {
		return nil
	}

	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	word, err := h.service.NextWord(
		c,
		ch.ID,
		cb.From.ID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotHost):
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.NotHost, nil),
				),
			)

		case errors.Is(err, ErrGameNotFound):
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.NotStarted, nil),
				),
			)

		default:
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.NextWord, nil),
				),
			)
		}
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(
			loc.T(
				i18n.Cmd.Crocodile.Callback.Word,
				i18n.CmdCrocodileCallbackWordData{
					Word: word,
				},
			),
		),
		botapi.WithCallbackAlert(),
	)
}

func (h *Handler) Finish(c *botapi.Context) error {
	cb := c.Update.CallbackQuery
	if cb == nil {
		return nil
	}

	cm := cctx.MustChatMember(c)
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if err := h.service.Stop(
		c,
		ch.ID,
		cm.ID(),
	); err != nil {
		switch {
		case errors.Is(err, ErrGameNotFound):
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.NotStarted, nil),
				),
			)

		case errors.Is(err, ErrNotHost):
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.NotHost, nil),
				),
			)

		default:
			_ = c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(i18n.Cmd.Crocodile.Error.Finish, nil),
				),
			)
			return err
		}
	}

	chatID, _ := c.Chat()

	_, _ = c.Bot.EditMessageText(
		c,
		chatID,
		cb.Message.MessageID,
		loc.T(i18n.Cmd.Crocodile.Callback.Finished, nil),
		botapi.WithReplyMarkup(becomeHostKeyboard(loc)),
	)

	return c.AnswerCallback()
}

func (h *Handler) Stop(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)

	err := h.service.Stop(
		c,
		ch.ID,
		cm.ID(),
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrGameNotFound):
			_, err = c.Reply(
				loc.T(i18n.Cmd.Crocodile.Error.NotStarted, nil),
			)

		case errors.Is(err, ErrNotHost):
			_, err = c.Reply(
				loc.T(i18n.Cmd.Crocodile.Error.NotHost, nil),
			)

		default:
			_, err = c.Reply(
				loc.T(i18n.Cmd.Crocodile.Error.Finish, nil),
			)
		}

		return err
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.Crocodile.Stopped, nil),
	)

	return err
}

func (h *Handler) HandleMessage(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}

	ch := cctx.MustChat(c)
	sender := cctx.MustChatMember(c)

	game, guessed, err := h.service.Guess(
		c,
		ch.ID,
		sender.ID(),
		msg.Text,
	)

	if err != nil {
		if errors.Is(err, ErrGameNotFound) {
			return nil
		}

		return err
	}

	if !guessed {
		return nil
	}

	loc := cctx.MustLocalizer(c)

	mention := tghtml.MemberMention(
		loc,
		ch,
		sender,
	)

	_, err = c.Reply(
		loc.T(
			i18n.Cmd.Crocodile.Winner,
			i18n.CmdCrocodileWinnerData{
				User: mention,
				Word: game.Word,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(becomeHostKeyboard(loc)),
	)
	return err
}

func becomeHostKeyboard(loc *i18n.Localizer) *botapi.InlineKeyboardMarkup {
	return botapi.InlineKeyboard(
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.BecomeHost, nil),
				callbackBecomeHost,
			),
		),
	)
}

func (h *Handler) BecomeHost(c *botapi.Context) error {
	cb := c.Update.CallbackQuery
	if cb == nil {
		return nil
	}

	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	cm := cctx.MustChatMember(c)

	_, err := h.service.Start(
		c,
		ch.ID,
		cb.From.ID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrGameAlreadyExists):
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(
						i18n.Cmd.Crocodile.Error.AlreadyStarted,
						nil,
					),
				),
			)

		default:
			return c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(
						i18n.Cmd.Crocodile.Error.Start,
						nil,
					),
				),
			)
		}
	}

	mention := tghtml.MemberMention(
		loc,
		ch,
		cm,
	)

	chatID, _ := c.Chat()
	_, err = c.Bot.EditMessageText(
		c,
		chatID,
		cb.Message.MessageID,
		cb.Message.OriginalTextHTML()+"\n\n"+loc.T(
			i18n.Cmd.Crocodile.Continued,
			i18n.CmdCrocodileContinuedData{
				User: mention,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	if err != nil {
		return err
	}

	keyboard := botapi.InlineKeyboard(
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.ShowWord, nil),
				callbackShowWord,
			),
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.Next, nil),
				callbackNextWord,
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Crocodile.Button.Finish, nil),
				callbackFinish,
			),
		),
	)

	_, err = c.Bot.SendMessage(
		c,
		chatID,
		loc.T(
			i18n.Cmd.Crocodile.Started,
			i18n.CmdCrocodileStartedData{
				Host: mention,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(keyboard),
	)

	if err != nil {
		return err
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(
			loc.T(i18n.Cmd.Crocodile.Callback.Host, nil),
		),
	)
}

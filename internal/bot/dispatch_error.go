package bot

import (
	"activity-bot/internal/logger"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func telegramDispatchErrorHandler(_ *ext.Context, u *ext.Update, errMsg string) error {
	chatID := u.EffectiveChat().GetID()

	args := []any{
		"err", errMsg,
		"chat_id", chatID,
	}
	if usr := u.EffectiveUser(); usr != nil {
		args = append(args, "user_id", usr.ID)
		if usr.Username != "" {
			args = append(args, "username", usr.Username)
		}
	}

	if m := u.EffectiveMessage; m != nil {
		args = append(args,
			"message_id", m.ID,
			"message_text", logger.Truncate(m.Text, 2048),
			"is_service", m.IsService,
		)
		if m.ReplyTo != nil {
			if rt, ok := m.ReplyTo.(*tg.MessageReplyHeader); ok && rt.ReplyToMsgID != 0 {
				args = append(args, "reply_to_msg_id", rt.ReplyToMsgID)
			}
		}
	}

	if cb := u.CallbackQuery; cb != nil {
		args = append(args,
			"callback_msg_id", cb.MsgID,
			"callback_query_id", cb.QueryID,
			"callback_data", logger.Truncate(string(cb.Data), 512),
		)
	}

	logger.L.Info("update handler error", args...)
	return dispatcher.ContinueGroups
}

package events

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/participant"
	"activity-bot/internal/utils/tghtml"
	"context"
	"fmt"

	"github.com/gotd/botapi"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

func (h *Handler) ParticipantLeft(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant) error {
	lang := "ru"
	if u.NewParticipant != nil {
		return nil
	}

	var chatID constant.TDLibPeerID
	chatID.Channel(u.ChannelID)

	name := h.translator.T(lang, i18n.UserUnknown, nil)
	if user, ok := e.Users[u.UserID]; ok {
		name = user.FirstName
	}

	tag := participant.Rank(u.PrevParticipant)
	if tag == "" {
		tag = name
	}

	text := h.translator.T(lang, i18n.UserLeftMale, map[string]any{
		"User": tghtml.Mention(participant.ID(u.PrevParticipant), tag),
	})

	_, err := h.bot.SendMessage(ctx, botapi.ID(int64(chatID)),
		fmt.Sprintf(text),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"activity-bot/internal/i18n"
	"activity-bot/internal/user"
	"activity-bot/internal/utils/chatmembers"
	"activity-bot/internal/utils/participant"
	"activity-bot/internal/utils/tghtml"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"

	"github.com/gotd/botapi"
)

func (h *Handler) ParticipantUpdate(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant) error {
	log.Printf("participant %+v\n", u)

	var peerID constant.TDLibPeerID

	peerID.Channel(u.ChannelID)

	chatID := int64(peerID)

	if u.NewParticipant == nil || participant.IsBanned(u.NewParticipant) {
		return h.processLeft(ctx, u, chatID)
	}

	if u.PrevParticipant == nil || participant.IsBanned(u.PrevParticipant) {
		return h.processJoin(ctx, e, u, chatID)
	}

	newTag := participant.Rank(u.NewParticipant)
	if participant.Rank(u.PrevParticipant) != newTag {
		return h.memberService.UpdateTag(ctx, chatID, u.UserID, newTag)
	}

	return nil
}
func (h *Handler) processJoin(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant, chatID int64) error {
	ue := e.Users[u.UserID]

	userDTO := user.New(u.UserID, ue.FirstName, ue.LastName, ue.Username, user.GenderUnknown, ue.Bot, time.Now())

	res, err := h.memberService.HandleJoin(
		ctx,
		chatID,
		e.Channels[u.ChannelID].Title,
		userDTO,
		participant.Rank(u.NewParticipant),
	)
	if err != nil {
		return fmt.Errorf("handler process join: %w", err)
	}

	loc := h.translator.Localizer(res.ChatMember.Chat.Lang)

	if res.IsNew && participant.IsSelf(u.NewParticipant) {
		members, err := participant.GetChatMembers(h.bot, ctx, e, u.ChannelID)
		if err != nil {
			return fmt.Errorf("get chat members on bot join: %w", err)
		}

		if err = h.memberService.SyncChatMembers(
			ctx,
			chatID,
			chatmembers.ExtractMembers(members),
		); err != nil {
			return fmt.Errorf("sync chat members: %w", err)
		}

		var text string

		if participant.IsAdmin(u.NewParticipant) {
			text = loc.T(i18n.System.BotAddedAdmin, nil)
		} else {
			text = loc.T(i18n.System.BotAdded, nil)
		}

		_, err = h.bot.SendMessage(
			ctx,
			botapi.ID(chatID),
			text,
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	cm := res.ChatMember

	mention := tghtml.MemberMention(loc, cm.Chat, cm)

	key := i18n.User.Returned
	data := i18n.UserReturnedMaleData{
		User: mention,
	}

	if res.IsNew {
		key = i18n.User.Joined
	}

	text := loc.TGender(
		cm.User.Gender,
		key,
		data,
	)

	_, err = h.bot.SendMessage(
		ctx,
		botapi.ID(chatID),
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) processLeft(ctx context.Context, u *tg.UpdateChannelParticipant, chatID int64) error {
	cm, err := h.memberService.HandleLeft(ctx, chatID, u.UserID)
	if err != nil {
		return fmt.Errorf("handler process left: %w", err)
	}

	loc := h.translator.Localizer(cm.Chat.Lang)

	mention := tghtml.MemberMention(loc, cm.Chat, cm)

	text := loc.TGender(
		cm.User.Gender,
		i18n.User.Left,
		i18n.UserLeftMaleData{
			User: mention,
		},
	)

	_, err = h.bot.SendMessage(
		ctx,
		botapi.ID(chatID),
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

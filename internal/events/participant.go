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
	channel, ok := e.Channels[u.ChannelID]
	if !ok {
		return nil
	}

	if !channel.Megagroup {
		return nil
	}
	log.Printf("participant %+v\n", u)

	var peerID constant.TDLibPeerID
	peerID.Channel(u.ChannelID)

	chatID := int64(peerID)

	if participant.IsNotInChat(u.NewParticipant) &&
		!participant.IsNotInChat(u.PrevParticipant) {
		return h.processLeft(ctx, u, chatID)
	}

	if participant.IsNotInChat(u.PrevParticipant) &&
		!participant.IsNotInChat(u.NewParticipant) {
		return h.processJoin(ctx, e, u, chatID)
	}

	newTag := participant.Rank(u.NewParticipant)
	if participant.Rank(u.PrevParticipant) != newTag {
		if err := h.memberService.UpdateTag(ctx, chatID, u.UserID, newTag); err != nil {
			return fmt.Errorf("participant update tag: %w", err)
		}
		if err := h.roleUpdater.Update(ctx, chatID, h.bot); err != nil {
			return fmt.Errorf("participant update: %w", err)
		}
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
			text = loc.T(i18n.System.BotAddedAdmin, i18n.SystemBotAddedData{
				Emoji: tghtml.PatPatEmoji(),
			})
		} else {
			text = loc.T(i18n.System.BotAdded, i18n.SystemBotAddedAdminData{
				Emoji: tghtml.PatPatEmoji(),
			})
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

	var key i18n.MessageID
	var data any

	if res.IsNew {
		key = i18n.User.Joined
		data = i18n.UserJoinedData{
			User: mention,
		}
	} else {
		key = i18n.User.Returned
		data = i18n.UserReturnedData{
			User: mention,
		}
	}

	text := loc.T(
		key,
		data,
		i18n.WithGender(cm.Gender()),
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

	memberLink := tghtml.MemberLink(loc, cm.Chat, cm)

	text := loc.T(
		i18n.User.Left,
		i18n.UserLeftData{
			User: memberLink,
		},
		i18n.WithGender(cm.Gender()),
	)

	_, err = h.bot.SendMessage(
		ctx,
		botapi.ID(chatID),
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

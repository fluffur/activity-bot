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
		return h.processLeft(ctx, e, u, chatID)
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

	if res.IsNew && participant.IsSelf(u.NewParticipant) {
		members, err := participant.GetChatMembers(h.bot, ctx, e, u.ChannelID)
		if err != nil {
			return fmt.Errorf("get chat members on bot join: %w", err)
		}

		if err = h.memberService.SyncChatMembers(ctx, chatID, chatmembers.ExtractMembers(members)); err != nil {
			return fmt.Errorf("sync chat members: %w", err)
		}

		text := h.translator.TIf(
			res.ChatMember.Chat.Lang,
			participant.IsAdmin(u.NewParticipant),
			i18n.System.BotAddedAdmin,
			i18n.System.BotAdded,
			nil,
			nil,
		)

		_, err = h.bot.SendMessage(ctx, botapi.ID(chatID), text, botapi.WithParseMode(botapi.ParseModeHTML))

		return err
	}

	cm := res.ChatMember
	us := cm.User
	ch := cm.Chat

	mention := tghtml.UserMention(us.ID, cm.Display(h.translator.T(ch.Lang, i18n.User.Unknown), ch.EmojisEnabled))

	keyFemale, keyMale := i18n.User.ReturnedFemale, i18n.User.ReturnedMale
	argsFemale, argsMale := i18n.UserReturnedFemaleArgs(mention), i18n.UserReturnedMaleArgs(mention)

	if res.IsNew {
		keyFemale, keyMale = i18n.User.JoinedFemale, i18n.User.JoinedMale
		argsFemale, argsMale = i18n.UserJoinedFemaleArgs(mention), i18n.UserJoinedMaleArgs(mention)
	}

	text := h.translator.TIf(ch.Lang, us.Gender == user.GenderFemale, keyFemale, keyMale, argsFemale, argsMale)

	_, err = h.bot.SendMessage(ctx, botapi.ID(chatID), text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) processLeft(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant, chatID int64) error {
	cm, err := h.memberService.HandleLeft(ctx, chatID, u.UserID)
	if err != nil {
		return fmt.Errorf("handler process left: %w", err)
	}

	us := cm.User
	ch := cm.Chat

	mention := tghtml.UserMention(us.ID, cm.Display(h.translator.T(ch.Lang, i18n.User.Unknown), ch.EmojisEnabled))

	text := h.translator.TIf(
		ch.Lang,
		us.Gender == user.GenderFemale,
		i18n.User.LeftFemale,
		i18n.User.LeftMale,
		i18n.UserLeftFemaleArgs(mention),
		i18n.UserLeftMaleArgs(mention),
	)

	_, err = h.bot.SendMessage(ctx, botapi.ID(chatID), text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

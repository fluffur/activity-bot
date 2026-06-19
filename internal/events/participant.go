package events

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/user"
	"activity-bot/internal/utils/chatmembers"
	"activity-bot/internal/utils/participant"
	"activity-bot/internal/utils/tghtml"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	glog "github.com/gotd/log"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/jackc/pgx/v5"

	"github.com/gotd/botapi"
)

func (h *Handler) ParticipantUpdate(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant) error {
	log.Printf("participant %+v\n", u)

	var chatID constant.TDLibPeerID
	chatID.Channel(u.ChannelID)

	chatt, err := h.chatRepository.Get(ctx, int64(chatID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("participant update get chat: %w", err)
		}
		chatt = chat.New(int64(chatID), e.Channels[u.ChannelID].Title)

		if err := h.chatRepository.Create(ctx, chatt); err != nil {
			return fmt.Errorf("participant update create chat: %w", err)
		}
	}

	userr, err := h.userRepository.Get(ctx, u.UserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("join participant get user: %w", err)
		}
		userEntity := e.Users[u.UserID]
		userr = user.New(
			u.UserID,
			userEntity.FirstName,
			userEntity.LastName,
			userEntity.Username,
			user.GenderUnknown,
			userEntity.Bot,
			time.Now(),
		)

		if err := h.userRepository.Create(ctx, userr); err != nil {
			return fmt.Errorf("join participant create user: %w", err)
		}
	}

	if u.PrevParticipant == nil || participant.IsBanned(u.PrevParticipant) {
		return h.handleParticipantJoin(ctx, e, u, userr, chatt)
	}
	if u.NewParticipant == nil || participant.IsBanned(u.NewParticipant) {
		return h.handleParticipantLeft(ctx, e, u, userr, chatt)
	}

	newRank := participant.Rank(u.NewParticipant)
	if participant.Rank(u.PrevParticipant) != newRank {
		return h.chatMemberRepository.SetTag(ctx, int64(chatID), u.UserID, newRank)
	}

	return nil
}

func (h *Handler) handleParticipantLeft(
	ctx context.Context,
	e tg.Entities,
	u *tg.UpdateChannelParticipant,
	userr user.User,
	chatt chat.Chat,
) error {
	if err := h.chatMemberRepository.MarkLeft(ctx, chatt.ID, u.UserID, time.Now()); err != nil {
		glog.For(h.bot.Logger()).Warn(ctx, "Failed to mark participant as left", glog.Error(err))
	}

	name := h.translator.T(chatt.Language, i18n.UserUnknown, nil)
	if userEntity, ok := e.Users[u.UserID]; ok {
		name = userEntity.FirstName
	}

	tag := participant.Rank(u.PrevParticipant)
	if tag == "" {
		tag = name
	}

	mention := tghtml.UserMention(userr.ID, tag)

	text := h.translator.TIf(
		chatt.Language,
		userr.Gender == user.GenderFemale,
		i18n.UserLeftFemale,
		i18n.UserLeftMale,
		i18n.UserLeftFemaleArgs(mention),
		i18n.UserLeftMaleArgs(mention),
	)

	_, err := h.bot.SendMessage(ctx, botapi.ID(chatt.ID),
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) handleParticipantJoin(
	ctx context.Context,
	e tg.Entities,
	u *tg.UpdateChannelParticipant,
	userr user.User,
	chatt chat.Chat,
) error {
	newMember := false
	chatMember, err := h.chatMemberRepository.Get(ctx, chatt.ID, u.UserID)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("join participant get chat member: %w", err)
		}
		newMember = true
		chatMember = chatmember.New(
			userr,
			chatt,
			participant.Rank(u.NewParticipant),
			chatmember.StatusMember,
			time.Now(),
		)

		if err := h.chatMemberRepository.Create(ctx, chatMember); err != nil {
			return fmt.Errorf("join participant create chat member: %w", err)
		}
	}

	name := h.translator.T(chatt.Language, i18n.UserUnknown, nil)
	if userEntity, ok := e.Users[u.UserID]; ok {
		name = userEntity.FirstName
	}

	tag := chatMember.Tag
	if tag == "" {
		tag = name
	}
	if err := h.chatMemberRepository.MarkLeft(ctx, chatt.ID, u.UserID, time.Time{}); err != nil {
		glog.For(h.bot.Logger()).Warn(ctx, "Failed to mark participant as not left", glog.Error(err))
	}

	if newMember && participant.IsSelf(u.NewParticipant) {
		members, err := participant.GetChatMembers(h.bot, ctx, e, u)
		if err != nil {
			return fmt.Errorf("join participant get chat members: %w", err)
		}

		cms := chatmembers.ExtractMembers(members)

		cmIDs := make([]int64, 0, len(cms))
		for _, cm := range cms {
			cmIDs = append(cmIDs, cm.User.ID)
		}

		if err := h.chatMemberRepository.MarkAllLeftExcept(ctx, chatt.ID, cmIDs, time.Now()); err != nil {
			return fmt.Errorf("join sync mark all left: %w", err)
		}
		if err := h.chatMemberRepository.UpsertChatMembers(ctx, chatt.ID, cms); err != nil {
			return fmt.Errorf("join sync upsert chat members: %w", err)
		}

		text := h.translator.TIf(
			chatt.Language,
			participant.IsAdmin(u.NewParticipant),
			i18n.BotAddedAdmin,
			i18n.BotAdded,
			nil,
			nil,
		)

		if _, err := h.bot.SendMessage(ctx, botapi.ID(chatt.ID), text, botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
			return err
		}
		return nil
	}

	mention := tghtml.UserMention(participant.ID(u.NewParticipant), tag)

	if newMember {
		text := h.translator.TIf(
			chatt.Language,
			userr.Gender == user.GenderFemale,
			i18n.UserJoinedFemale,
			i18n.UserJoinedMale,
			i18n.UserJoinedFemaleArgs(mention),
			i18n.UserJoinedMaleArgs(mention),
		)
		if _, err := h.bot.SendMessage(ctx, botapi.ID(chatt.ID), text, botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
			return err
		}

		return nil
	}

	text := h.translator.TIf(
		chatt.Language,
		userr.Gender == user.GenderFemale,
		i18n.UserReturnedFemale,
		i18n.UserReturnedMale,
		i18n.UserReturnedFemaleArgs(mention),
		i18n.UserReturnedMaleArgs(mention),
	)

	if _, err := h.bot.SendMessage(ctx, botapi.ID(chatt.ID), text, botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
		return err
	}

	return nil
}

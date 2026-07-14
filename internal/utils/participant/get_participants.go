package participant

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/gotd/botapi"
)

func GetChatMembers(bot *botapi.Bot, ctx context.Context, e tg.Entities, channelID int64) ([]botapi.ChatMember, error) {
	channelID = BotAPIChatIDToChannelID(channelID)
	res, err := bot.Raw().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: e.Channels[channelID].AsInput(),
		Filter:  &tg.ChannelParticipantsRecent{},
		Limit:   200,
	})
	if err != nil {
		return nil, fmt.Errorf("sync get participants: %w", err)
	}

	participants, ok := res.AsModified()
	if !ok {
		return nil, fmt.Errorf("sync participants as modified")
	}

	users := usersByID(participants.Users)
	out := make([]botapi.ChatMember, 0, len(participants.Participants))

	for _, p := range participants.Participants {
		out = append(out, chatMemberFromParticipant(p, users))
	}

	return out, nil
}

func usersByID(users []tg.UserClass) map[int64]*tg.User {
	m := make(map[int64]*tg.User, len(users))
	for _, u := range users {
		if us, ok := u.(*tg.User); ok {
			m[us.ID] = us
		}
	}

	return m
}

func chatMemberFromParticipant(p tg.ChannelParticipantClass, users map[int64]*tg.User) botapi.ChatMember {
	u := func(id int64) botapi.User {
		if u, ok := users[id]; ok {
			return userFromTgUser(u)
		}

		return botapi.User{ID: id}
	}

	switch v := p.(type) {
	case *tg.ChannelParticipantCreator:
		return &botapi.ChatMemberOwner{
			Status:      botapi.StatusCreator,
			User:        u(v.UserID),
			IsAnonymous: v.AdminRights.Anonymous,
			CustomTitle: v.Rank,
		}
	case *tg.ChannelParticipantAdmin:
		r := v.AdminRights

		return &botapi.ChatMemberAdministrator{
			Status:              botapi.StatusAdministrator,
			User:                u(v.UserID),
			CanBeEdited:         v.CanEdit,
			IsAnonymous:         r.Anonymous,
			CanManageChat:       r.Other,
			CanDeleteMessages:   r.DeleteMessages,
			CanManageVideoChats: r.ManageCall,
			CanRestrictMembers:  r.BanUsers,
			CanPromoteMembers:   r.AddAdmins,
			CanChangeInfo:       r.ChangeInfo,
			CanInviteUsers:      r.InviteUsers,
			CanPostMessages:     r.PostMessages,
			CanEditMessages:     r.EditMessages,
			CanPinMessages:      r.PinMessages,
			CustomTitle:         v.Rank,
		}
	case *tg.ChannelParticipantSelf:
		return &botapi.ChatMemberMember{Status: botapi.StatusMember, User: u(v.UserID), Tag: v.Rank}
	case *tg.ChannelParticipant:
		return &botapi.ChatMemberMember{Status: botapi.StatusMember, User: u(v.UserID), Tag: v.Rank}
	case *tg.ChannelParticipantBanned:
		uid := peerUserID(v.Peer)
		if v.Left {
			return &botapi.ChatMemberLeft{Status: botapi.StatusLeft, User: u(uid)}
		}

		br := v.BannedRights
		// A member who can still view messages is "restricted"; one who cannot is
		// fully banned ("kicked").
		if br.ViewMessages {
			return &botapi.ChatMemberBanned{
				Status:    botapi.StatusBanned,
				User:      u(uid),
				UntilDate: br.UntilDate,
			}
		}

		return &botapi.ChatMemberRestricted{
			Status:                botapi.StatusRestricted,
			User:                  u(uid),
			IsMember:              !v.Left,
			CanSendMessages:       !br.SendMessages,
			CanSendMediaMessages:  !br.SendMedia,
			CanSendPolls:          !br.SendPolls,
			CanSendOtherMessages:  !(br.SendStickers || br.SendGifs || br.SendGames || br.SendInline),
			CanAddWebPagePreviews: !br.EmbedLinks,
			CanChangeInfo:         !br.ChangeInfo,
			CanInviteUsers:        !br.InviteUsers,
			CanPinMessages:        !br.PinMessages,
			Tag:                   v.Rank,
			UntilDate:             br.UntilDate,
		}
	case *tg.ChannelParticipantLeft:
		return &botapi.ChatMemberLeft{Status: botapi.StatusLeft, User: u(peerUserID(v.Peer))}
	default:
		return &botapi.ChatMemberMember{Status: botapi.StatusMember}
	}
}

func userFromTgUser(u *tg.User) botapi.User {
	return botapi.User{
		ID:           u.ID,
		IsBot:        u.Bot,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     u.Username,
		LanguageCode: u.LangCode,
		IsPremium:    u.Premium,
	}
}

// peerUserID extracts the user id from a peer, or 0 if it is not a user.
func peerUserID(p tg.PeerClass) int64 {
	if u, ok := p.(*tg.PeerUser); ok {
		return u.UserID
	}

	return 0
}

func BotAPIChatIDToChannelID(chatID int64) int64 {
	if chatID >= 0 {
		return chatID
	}

	return -(chatID + 1000000000000)
}

package adapter

import (
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/gotd/td/tg"
)

type TelegramChatMembersProvider struct {
	client *gotgproto.Client
}

func NewTelegramChatMembersProvider(client *gotgproto.Client) *TelegramChatMembersProvider {
	return &TelegramChatMembersProvider{
		client: client,
	}
}

func (p *TelegramChatMembersProvider) GetChatMembers(ctx context.Context, chatID int64) ([]model.ChatMemberUpdate, error) {
	chat, err := p.resolveChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	switch ch := chat.(type) {
	case *tg.Channel:
		return p.getChannelMembers(ctx, ch)

	case *tg.Chat:
		return p.getBasicChatMembers(ctx, ch)

	default:
		return nil, errors.New("unsupported chat type")
	}
}

func (p *TelegramChatMembersProvider) resolveChat(ctx context.Context, chatID int64) (tg.ChatClass, error) {

	peerID := chatID
	if peerID < 0 {
		peerID = -peerID
		if peerID > 1000000000000 {
			peerID -= 1000000000000
		}
	}
	chRes, err := p.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: peerID},
	})
	if err == nil && len(chRes.GetChats()) > 0 {
		return chRes.GetChats()[0], nil
	}
	logger.L.Warn("no channel found", "error", err)
	chatRes, err := p.client.API().MessagesGetChats(ctx, []int64{chatID})
	if err != nil {
		return nil, err
	}

	if len(chatRes.GetChats()) == 0 {
		return nil, errors.New("chat not found")
	}

	return chatRes.GetChats()[0], nil
}

func (p *TelegramChatMembersProvider) getChannelMembers(
	ctx context.Context,
	fullChannel *tg.Channel,
) ([]model.ChatMemberUpdate, error) {

	var result []model.ChatMemberUpdate

	offset := 0
	limit := 200

	for {
		c, err := p.client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: &tg.InputChannel{
				ChannelID:  fullChannel.ID,
				AccessHash: fullChannel.AccessHash,
			},
			Filter: &tg.ChannelParticipantsRecent{},
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}

		participants, ok := c.(*tg.ChannelsChannelParticipants)
		if !ok {
			break
		}

		if len(participants.Participants) == 0 {
			break
		}

		userMap := make(map[int64]*tg.User)
		for _, u := range participants.Users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		for _, participant := range participants.Participants {
			var userID int64
			var status int16
			var tag string

			switch p := participant.(type) {
			case *tg.ChannelParticipant:
				userID = p.UserID
				tag = p.Rank

			case *tg.ChannelParticipantSelf:
				userID = p.UserID
				tag = p.Rank

			case *tg.ChannelParticipantCreator:
				userID = p.UserID
				tag = p.Rank
				status = 5

			case *tg.ChannelParticipantAdmin:
				userID = p.UserID
				tag = p.Rank

			case *tg.ChannelParticipantBanned:
				peer := p.GetPeer()
				u, ok := peer.(*tg.PeerUser)
				if !ok {
					continue
				}
				userID = u.UserID
				tag = p.Rank

			default:
				continue
			}

			u, ok := userMap[userID]
			if !ok || u.Bot {
				continue
			}

			result = append(result, model.ChatMemberUpdate{
				User: model.User{
					ID:        u.ID,
					FirstName: u.FirstName,
					LastName:  u.LastName,
					Username:  u.Username,
				},
				Tag:    tag,
				Status: status,
			})
		}

		offset += len(participants.Participants)
		if offset >= participants.Count {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return result, nil
}

func (p *TelegramChatMembersProvider) getBasicChatMembers(ctx context.Context, ch *tg.Chat) ([]model.ChatMemberUpdate, error) {
	full, err := p.client.API().MessagesGetFullChat(ctx, ch.ID)
	if err != nil {
		return nil, err
	}

	fullChat := full.FullChat.(*tg.ChatFull)

	participants, ok := fullChat.Participants.(*tg.ChatParticipants)
	if !ok {
		return nil, errors.New("invalid chat type")
	}

	userMap := make(map[int64]*tg.User)
	for _, u := range full.Users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	var result []model.ChatMemberUpdate

	for _, p := range participants.Participants {
		var userID int64
		var status int16
		var tag string

		switch cp := p.(type) {
		case *tg.ChatParticipantCreator:
			userID = cp.UserID
			tag = cp.Rank
			status = 5

		case *tg.ChatParticipantAdmin:
			userID = cp.UserID
			tag = cp.Rank

		case *tg.ChatParticipant:
			userID = cp.UserID

		default:
			continue
		}

		u, ok := userMap[userID]
		if !ok || u.Bot {
			continue
		}

		result = append(result, model.ChatMemberUpdate{
			User: model.User{
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Username:  u.Username,
			},
			Tag:    tag,
			Status: status,
		})
	}

	return result, nil
}

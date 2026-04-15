package adapter

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"log"
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
	var result []model.ChatMemberUpdate

	peerID := chatID
	if peerID < 0 {
		peerID = -peerID
		if peerID > 1000000000000 {
			peerID -= 1000000000000
		}
	}

	d, err := p.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{&tg.InputChannel{ChannelID: peerID}})
	if err != nil {
		return nil, err
	}
	chats := d.GetChats()
	if len(chats) == 0 {
		log.Println(chatID, peerID)
		c, err := p.client.API().MessagesGetChats(ctx, []int64{chatID})
		if err != nil {
			return nil, err
		}
		chats = c.GetChats()
		if len(chats) == 0 {
			return nil, errors.New("no chats found")
		}
	}
	ch := chats[0]
	fullChannel, ok := ch.(*tg.Channel)
	if !ok {
		return nil, errors.New("not a channel")
	}

	offset := 0
	limit := 200

	for {
		c, err := p.client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: &tg.InputChannel{
				ChannelID:  fullChannel.ID,
				AccessHash: fullChannel.AccessHash,
			},
			Filter: &tg.ChannelParticipantsSearch{},
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

		time.Sleep(time.Millisecond * 500)
	}

	return result, nil
}

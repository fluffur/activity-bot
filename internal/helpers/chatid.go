package helpers

import "github.com/gotd/td/constant"

func ChannelChatID(channelID int64) int64 {
	var id constant.TDLibPeerID
	id.Channel(channelID)
	return int64(id)
}

func BasicChatID(chatID int64) int64 {
	var id constant.TDLibPeerID
	id.Chat(chatID)
	return int64(id)
}

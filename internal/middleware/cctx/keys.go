package cctx

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"context"
	"fmt"
)

type ChatMemberKey struct{}
type ChatKey struct{}

func Chat(ctx context.Context) (chat.Chat, error) {
	m, ok := ctx.Value(ChatKey{}).(chat.Chat)
	if !ok {
		return chat.Chat{}, fmt.Errorf("chat not found")
	}

	return m, nil
}

func ChatMember(ctx context.Context) (chatmember.ChatMember, error) {
	m, ok := ctx.Value(ChatMemberKey{}).(chatmember.ChatMember)
	if !ok {
		return chatmember.ChatMember{}, fmt.Errorf("chat member not found")
	}

	return m, nil
}

package cctx

import (
	"activity-bot/internal/chatmember"
	"context"
	"fmt"
)

type ChatMemberKey struct{}

func ChatMember(ctx context.Context) (chatmember.ChatMember, error) {
	m, ok := ctx.Value(ChatMemberKey{}).(chatmember.ChatMember)
	if !ok {
		return chatmember.ChatMember{}, fmt.Errorf("chat member not found")
	}

	return m, nil
}

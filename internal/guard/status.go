package guard

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"context"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func Restricted(memberService *member.Service, chatService *chat.Service, sessionService cmd.SessionService, commandName string, defaultStatus model.Status) cmd.Guard {
	return cmd.GuardFunc(func(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
		chatID, err := cmd.GetChatID(sessionService, ctx, stdCtx)
		if err != nil {
			return false, "Не удалось определить чат"
		}

		required := defaultStatus
		if chatService != nil && commandName != "" {
			if s, err := chatService.GetCommandPermission(stdCtx, chatID, commandName); err == nil {
				required = s
			}
		}

		m, err := memberService.GetChatMember(stdCtx, chatID, ctx.EffectiveUser.Id)
		if err != nil {
			return false, "Ошибка при проверке прав"
		}

		if m.Status < required {
			return false, fmt.Sprintf("%s Требуются права: %s", helpers.StatusEmojiPlain(required), required.String())
		}

		return true, ""
	})
}

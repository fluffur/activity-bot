package helpers

import (
	"activity-bot/internal/model"
	"context"
	"strings"

	"github.com/go-telegram/bot/models"
)

type UserService interface {
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
}

func ExtractTargetUser(ctx context.Context, userService UserService, update *models.Update, args string) (int64, string, bool, error) {
	var userID *int64

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		userID = &update.Message.ReplyToMessage.From.ID
	} else if update.Message.ExternalReply != nil && update.Message.ExternalReply.Origin.MessageOriginUser != nil {
		userID = &update.Message.ExternalReply.Origin.MessageOriginUser.SenderUser.ID
	}

	if userID != nil {
		return *userID, args, true, nil
	}

	textRunes := []rune(update.Message.Text)
	for _, e := range update.Message.Entities {
		if e.User != nil {
			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))
			return e.User.ID, restArg, true, nil
		} else if e.Type == "mention" {
			username := textRunes[e.Offset : e.Offset+e.Length]
			u, err := userService.GetUserByUsername(ctx, string(username[1:]))
			if err != nil {
				return 0, "", false, err
			}
			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))
			return u.ID, restArg, true, nil
		}
	}

	return 0, args, false, nil
}

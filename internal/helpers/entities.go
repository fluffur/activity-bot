package helpers

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"strings"

	"github.com/go-telegram/bot/models"
)

var ErrUserNotSpecified = errors.New("user not specified")

type UserService interface {
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetUser(ctx context.Context, id int64) (model.User, error)
}

func ExtractTargetUser(
	ctx context.Context,
	userService UserService,
	update *models.Update,
	args string,
) (model.User, string, error) {

	var tgUserID *int64

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		tgUserID = &update.Message.ReplyToMessage.From.ID
	} else if update.Message.ExternalReply != nil &&
		update.Message.ExternalReply.Origin.MessageOriginUser != nil {
		tgUserID = &update.Message.ExternalReply.Origin.MessageOriginUser.SenderUser.ID
	}

	if tgUserID != nil {
		u, err := userService.GetUser(ctx, *tgUserID)
		if err != nil {
			return model.User{}, "", err
		}
		return u, args, nil
	}

	textRunes := []rune(update.Message.Text)

	for _, e := range update.Message.Entities {

		if e.User != nil {
			u, err := userService.GetUser(ctx, e.User.ID)
			if err != nil {
				return model.User{}, "", err
			}

			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))

			return u, restArg, nil
		}

		if e.Type == "mention" {
			username := string(textRunes[e.Offset+1 : e.Offset+e.Length])

			u, err := userService.GetUserByUsername(ctx, username)
			if err != nil {
				return model.User{}, "", err
			}

			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))

			return u, restArg, nil
		}
	}

	return model.User{}, args, ErrUserNotSpecified
}

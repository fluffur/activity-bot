package middleware

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/user"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gotd/botapi"
	"github.com/jackc/pgx/v5"
)

func ChatMiddleware(
	chatRepository chat.Repository,
	userRepository user.Repository,
	chatMemberRepository chatmember.Repository,
) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {
			msg := c.Message()
			if msg == nil {
				msg = c.Update.CallbackQuery.Message
			}
			if msg == nil {
				return next(c)
			}
			if msg.Chat.Type != botapi.ChatTypeGroup && msg.Chat.Type != botapi.ChatTypeSupergroup {
				// add pm sessions later
				return next(c)
			}

			chatID, ok := c.Chat()
			if !ok {
				return nil
			}

			id, ok := chatID.(botapi.ChatIDInt)
			if !ok {
				return nil
			}

			ctx := c.Context

			chatModel, err := getOrCreateChat(ctx, chatRepository, c.Bot, int64(id))
			if err != nil {
				return err
			}

			sender := c.Sender()
			if sender == nil {
				return fmt.Errorf("sender is nil")
			}

			userModel, err := getOrCreateUser(ctx, userRepository, sender)
			if err != nil {
				return err
			}

			member, err := getOrCreateChatMember(
				ctx,
				chatMemberRepository,
				c.Bot,
				chatModel,
				userModel,
				int64(id),
				sender.ID,
			)
			if err != nil {
				return err
			}

			ctx = context.WithValue(ctx, cctx.ChatMemberKey{}, member)
			c.Context = ctx
			return next(c)
		}
	}
}

func getOrCreateChat(
	ctx context.Context,
	repo chat.Repository,
	bot *botapi.Bot,
	id int64,
) (chat.Chat, error) {

	model, err := repo.GetByID(ctx, id)
	if err == nil {
		return model, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return chat.Chat{}, fmt.Errorf("get chat: %w", err)
	}

	info, err := bot.GetChat(ctx, botapi.ChatIDInt(id))
	if err != nil {
		return chat.Chat{}, fmt.Errorf("get chat info: %w", err)
	}

	model = chat.New(id, info.Title)

	if err := repo.Create(ctx, model); err != nil {
		return chat.Chat{}, fmt.Errorf("create chat: %w", err)
	}

	return model, nil
}

func getOrCreateUser(
	ctx context.Context,
	repo user.Repository,
	sender *botapi.User,
) (user.User, error) {

	model, err := repo.GetByID(ctx, sender.ID)
	if err == nil {
		return model, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, fmt.Errorf("get user: %w", err)
	}

	model = user.New(
		sender.ID,
		sender.FirstName,
		sender.LastName,
		sender.Username,
		user.GenderUnknown,
		sender.IsBot,
		time.Now(),
	)

	if err := repo.Create(ctx, model); err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}

	return model, nil
}

func getOrCreateChatMember(
	ctx context.Context,
	repo chatmember.Repository,
	bot *botapi.Bot,
	chatModel chat.Chat,
	userModel user.User,
	chatID int64,
	userID int64,
) (chatmember.ChatMember, error) {

	member, err := repo.Get(ctx, chatID, userID)
	if err == nil {
		return member, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return chatmember.ChatMember{}, fmt.Errorf("get chat member: %w", err)
	}

	cm, err := bot.GetChatMember(ctx, botapi.ChatIDInt(chatID), userID)
	if err != nil {
		return chatmember.ChatMember{}, fmt.Errorf("get chat member info: %w", err)
	}

	status, tag, err := parseChatMember(cm)
	if err != nil {
		return chatmember.ChatMember{}, err
	}

	member = chatmember.New(
		userModel,
		chatModel,
		tag,
		status,
		time.Now(),
	)

	if err := repo.Create(ctx, member); err != nil {
		return chatmember.ChatMember{}, fmt.Errorf("create chat member: %w", err)
	}

	return member, nil
}

func parseChatMember(cm botapi.ChatMember) (chatmember.Status, string, error) {
	switch v := cm.(type) {
	case *botapi.ChatMemberOwner:
		log.Println(v.CustomTitle)

		return chatmember.StatusOwner, v.CustomTitle, nil

	case *botapi.ChatMemberAdministrator:
		return chatmember.StatusMember, v.CustomTitle, nil

	case *botapi.ChatMemberMember:
		return chatmember.StatusMember, v.Tag, nil

	case *botapi.ChatMemberRestricted:
		return chatmember.StatusMember, v.Tag, nil

	case *botapi.ChatMemberLeft, *botapi.ChatMemberBanned:
		return chatmember.StatusMember, "", fmt.Errorf("user is not active chat member")

	default:
		return chatmember.StatusMember, "", fmt.Errorf("unknown chat member type")
	}
}

package middleware

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/pmsession"
	"activity-bot/internal/user"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/log"
	"github.com/jackc/pgx/v5"

	"github.com/gotd/botapi"
)

func ChatMiddleware(cr chat.Repository, sr pmsession.Repository) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {
			msg := c.Message()
			if msg == nil {
				msg = c.Update.CallbackQuery.Message
			}

			if msg == nil {
				return next(c)
			}

			ctx := c.Context
			chatModel := chat.New(0, "")

			if msg.Chat.Type != botapi.ChatTypeGroup && msg.Chat.Type != botapi.ChatTypeSupergroup {
				ch, err := sr.GetChat(ctx, msg.Chat.ID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						ctx = context.WithValue(ctx, cctx.ChatKey{}, chatModel)
						c.Context = ctx

						return next(c)
					}

					return fmt.Errorf("get chat: %w", err)
				}

				chatModel = ch
			} else {
				ch, err := getOrCreateChat(ctx, cr, msg.Chat)
				if err != nil {
					return err
				}

				chatModel = ch
			}

			ctx = context.WithValue(ctx, cctx.ChatKey{}, chatModel)
			c.Context = ctx

			return next(c)
		}
	}
}

type UsernameChangedNotifier interface {
	NotifyUsernameChanged(c *botapi.Context, oldUsername, newUsername string) error
}

func ChatMemberMiddleware(ur user.Repository, cmr chatmember.Repository, notifier UsernameChangedNotifier) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {
			ctx := c.Context

			sender := c.Sender()
			msg := c.Message()
			if msg == nil {
				msg = c.Update.CallbackQuery.Message
			}

			if msg == nil {
				return next(c)
			}

			userModel, userChanged, err := getOrCreateUser(ctx, ur, sender, msg.Chat)
			if err != nil {
				return fmt.Errorf("chat member middleware get user: %w", err)
			}

			chatModel, err := cctx.Chat(ctx)
			if err != nil {
				return fmt.Errorf("chat member middleware get chat: %w", err)
			}

			if chatModel.ID == 0 {
				return next(c)
			}

			member, err := getOrCreateChatMember(
				ctx,
				cmr,
				c.Bot,
				chatModel,
				userModel,
			)
			if err != nil {
				return fmt.Errorf("chat member middleware get chat member: %w", err)
			}

			ctx = context.WithValue(ctx, cctx.ChatMemberKey{}, member)
			c.Context = ctx

			if userChanged.NewUsername != "" {
				if err := notifier.NotifyUsernameChanged(c, userChanged.OldUsername, userChanged.NewUsername); err != nil {
					log.For(c.Bot.Logger()).Error(c.Context, "notify username changed", log.Error(err))

					return next(c)
				}
			}

			return next(c)
		}
	}
}

func SaveMessageMiddleware(messageRepository message.Repository) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {
			sender := c.Sender()
			msg := c.Message()

			if msg == nil || sender == nil {
				return next(c)
			}

			if msg.Chat.Type == botapi.ChatTypePrivate {
				return nil
			}

			if msg.From == nil || msg.From.IsBot {
				return nil
			}

			if err := messageRepository.Save(
				c.Context,
				msg.Chat.ID,
				sender.ID,
				int64(msg.MessageID),
			); err != nil {
				return fmt.Errorf("save message: %w", err)
			}

			return next(c)
		}
	}
}

func getOrCreateChat(ctx context.Context, repo chat.Repository, ch botapi.Chat) (chat.Chat, error) {
	id := ch.ID

	model, err := repo.Get(ctx, id)
	if err == nil {
		return model, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return chat.Chat{}, fmt.Errorf("get chat: %w", err)
	}

	model = chat.New(id, ch.Title)

	if err := repo.Create(ctx, model); err != nil {
		return chat.Chat{}, fmt.Errorf("create chat: %w", err)
	}

	return model, nil
}

type UserUpdate struct {
	OldUsername string
	NewUsername string
}

func getOrCreateUser(
	ctx context.Context,
	repo user.Repository,
	sender *botapi.User,
	c botapi.Chat,
) (model user.User, userUpdate UserUpdate, err error) {
	var (
		senderID                                        int64
		senderUsername, senderFirstName, senderLastName string
		senderIsBot                                     bool
	)

	switch {
	case sender != nil:
		senderID = sender.ID
		senderUsername = sender.Username
		senderFirstName = sender.FirstName
		senderLastName = sender.LastName
		senderIsBot = sender.IsBot

	case c.Type == botapi.ChatTypePrivate:
		senderID = c.ID
		senderUsername = c.Username
		senderFirstName = c.FirstName
		senderLastName = c.LastName
	default:
		return model, userUpdate, fmt.Errorf("no user")
	}

	model, err = repo.Get(ctx, senderID)
	if err == nil {
		userUpdate.OldUsername = model.Username

		if model.Username != senderUsername {
			userUpdate.NewUsername = senderUsername
			model.Username = senderUsername
		}

		if model.FirstName != senderFirstName {
			model.FirstName = senderFirstName
		}

		if model.LastName != senderLastName {
			model.LastName = senderLastName
		}

		if model.Username != senderUsername ||
			model.FirstName != senderFirstName ||
			model.LastName != senderLastName {
			if err := repo.Update(ctx, model); err != nil {
				return model, userUpdate, err
			}
		}

		return model, userUpdate, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model, userUpdate, fmt.Errorf("get user: %w", err)
	}

	model = user.New(
		senderID,
		senderFirstName,
		senderLastName,
		senderUsername,
		user.GenderUnknown,
		senderIsBot,
		time.Now(),
	)

	if err := repo.Create(ctx, model); err != nil {
		return model, userUpdate, fmt.Errorf("create user: %w", err)
	}

	return model, userUpdate, nil
}

func getOrCreateChatMember(
	ctx context.Context,
	repo chatmember.Repository,
	bot *botapi.Bot,
	chatModel chat.Chat,
	userModel user.User,
) (chatmember.ChatMember, error) {
	chatID := chatModel.ID
	userID := userModel.ID

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

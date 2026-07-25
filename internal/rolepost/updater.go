package rolepost

import (
	"activity-bot/internal/chatmember"
	"context"
	"fmt"

	"github.com/gotd/botapi"
)

type RoleUpdater struct {
	repo          chatmember.Repository
	targetChatID  int64
	rolesPostChat string
}

func NewRoleUpdater(
	repo chatmember.Repository,
	targetChatID int64,
) *RoleUpdater {
	return &RoleUpdater{
		repo:         repo,
		targetChatID: targetChatID,
	}
}

func (r *RoleUpdater) Update(c context.Context, chatID int64, bot *botapi.Bot) error {
	if r.targetChatID != chatID {
		return nil
	}
	members, err := r.repo.List(c, chatmember.Filter{
		ChatID: r.targetChatID,
		IsBot: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
		Left: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("get chat members: %w", err)
	}

	text, err := Render(BuildRoleStates(members))
	if err != nil {
		return fmt.Errorf("render role states: %w", err)
	}
	_, err = bot.EditMessageCaption(
		c,
		botapi.Username("H4venflood"),
		15,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	if err != nil {
		return fmt.Errorf("update roles: %w", err)
	}

	return nil
}

func (r *RoleUpdater) UpdateMembersCount(c context.Context, chatID int64, bot *botapi.Bot) error {
	if r.targetChatID != chatID {
		return nil
	}
	members, err := r.repo.List(c, chatmember.Filter{
		ChatID: r.targetChatID,
		IsBot: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
		Left: chatmember.OptionalBool{
			Bool:  false,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("get chat members: %w", err)
	}

	text, err := RenderApplication(len(members), 45)
	if err != nil {
		return fmt.Errorf("render application: %w", err)
	}
	_, err = bot.EditMessageCaption(
		c,
		botapi.Username("H4venflood"),
		18,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	if err != nil {
		return fmt.Errorf("update roles: %w", err)
	}

	return nil
}

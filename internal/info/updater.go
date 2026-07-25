package info

import (
	"activity-bot/internal/chatmember"
	"context"
	"fmt"

	"github.com/gotd/botapi"
)

type Updater struct {
	repo          chatmember.Repository
	targetChatID  int64
	rolesPostChat string
}

func NewUpdater(
	repo chatmember.Repository,
	targetChatID int64,
) *Updater {
	return &Updater{
		repo:         repo,
		targetChatID: targetChatID,
	}
}

func (r *Updater) UpdateRolesPost(c context.Context, chatID int64, bot *botapi.Bot) error {
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

func (r *Updater) UpdateApplyPost(c context.Context, chatID int64, bot *botapi.Bot) error {
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
		return fmt.Errorf("update apply: %w", err)
	}

	return nil
}

func (r *Updater) UpdateRestsPost(c context.Context, chatID int64, bot *botapi.Bot) error {
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

	text, err := RenderRests(BuildRestMembers(members))
	if err != nil {
		return fmt.Errorf("render rests: %w", err)
	}

	_, err = bot.EditMessageCaption(
		c,
		botapi.Username("H4venflood"),
		17,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	if err != nil {
		return fmt.Errorf("update rests: %w", err)
	}

	return nil
}

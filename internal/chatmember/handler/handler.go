package handler

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/emoji"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rolepost"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/chatmembers"
	"activity-bot/internal/utils/participant"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/gotd/botapi"
)

const CategoryChatMember command.Category = "chat_member"

type Handler struct {
	repo         chatmember.Repository
	service      *chatmember.Service
	rolesPost    string
	targetChatID int64
}

func NewHandler(repo chatmember.Repository, service *chatmember.Service, rolesPost string, targetChatID int64,
) *Handler {
	return &Handler{repo: repo, service: service, rolesPost: rolesPost, targetChatID: targetChatID}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"setchatemoji",
			h.SetEmoji,
			i18n.Cmd.ChatMember.SetEmoji.Desc,
			CategoryChatMember,
			option.WithAliases("значок", "чат эмодзи"),
			option.IgnorePermissionCheck(),
			option.WithRules(rule.User().Optional(), rule.Text()),
			option.WithPermission(permission.StatusModerator),
		),
		action.NewCommand(
			"chatemoji",
			h.ShowEmoji,
			i18n.Cmd.ChatMember.ShowEmoji.Desc,
			CategoryChatMember,
			option.WithAliases("значок", "чат эмодзи"),
			option.WithRules(rule.User().Optional()),
		),
		action.NewCommand(
			"delchatemoji",
			h.RemoveEmoji,
			i18n.Cmd.ChatMember.RemoveEmoji.Desc,
			CategoryChatMember,
			option.WithAliases("-значок", "-чат эмодзи"),
			option.WithRules(rule.User().Optional()),
			option.WithPermission(permission.StatusModerator),
			option.IgnorePermissionCheck(),
		),
		action.NewCommand(
			"update",
			h.UpdateChatMembers,
			i18n.Cmd.ChatMember.Update.Desc,
			CategoryChatMember,
			option.WithAliases("обновить чат", "чат обновить"),
		),
		action.NewCommand(
			"ship",
			h.Ship,
			i18n.Cmd.Ship.Desc,
			CategoryChatMember,
			option.WithAliases("шипперим рандом", "шипперим"),
		),
		action.NewCommand(
			"updateroles",
			h.UpdateRoles,
			"BETA",
			CategoryChatMember,
		),
	}
}

func (h *Handler) SetEmoji(c *botapi.Context) error {
	msg := cctx.MustArgsMessage(c)

	emojis := emoji.Extract(msg.OriginalTextHTML())
	if len(emojis) == 0 || len(emojis) > 3 {
		return nil
	}

	emojisString := strings.Join(emojis, "")

	moderator := cctx.MustChatMember(c)
	target, ok := cctx.MustArgs(c).User()

	if !ok || !moderator.Permitted(cctx.MustPermission(c)) {
		target = moderator
	}

	ch := cctx.MustChat(c)

	if err := h.repo.SetEmoji(c, ch.ID, target.ID(), emojisString); err != nil {
		return fmt.Errorf("set member emoji: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.ChatMember.SetEmoji.Success, i18n.CmdChatMemberSetEmojiSuccessData{
		Emoji: emojisString,
		User:  tghtml.MemberMentionCustom(loc, target, false),
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) ShowEmoji(c *botapi.Context) error {
	cm := cctx.MustChatMember(c)
	if u, ok := cctx.MustArgs(c).User(); ok {
		cm = u
	}

	emojis := emoji.Extract(cm.Emojis)
	loc := cctx.MustLocalizer(c)

	if len(emojis) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.ChatMember.ShowEmoji.NoEmoji, i18n.CmdChatMemberShowEmojiNoEmojiData{
			User: tghtml.MemberMentionCustom(loc, cm, false),
		}), botapi.WithParseMode(botapi.ParseModeHTML))

		return err
	}

	_, err := c.Reply(loc.T(i18n.Cmd.ChatMember.ShowEmoji.Success, i18n.CmdChatMemberShowEmojiSuccessData{
		User:  tghtml.MemberMentionCustom(loc, cm, false),
		Emoji: cm.Emojis,
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) UpdateChatMembers(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	members, err := participant.GetChatMembers(c.Bot, c, c.Update.Entities, ch.ID)
	if err != nil {
		return fmt.Errorf("get chat members on update: %w", err)
	}

	if err = h.service.SyncChatMembers(
		c,
		ch.ID,
		chatmembers.ExtractMembers(members),
	); err != nil {
		return fmt.Errorf("update chat members: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err = c.Reply(loc.T(i18n.Cmd.ChatMember.Update.Success, nil), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) UpdateRoles(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	if ch.ID != h.targetChatID {
		return nil
	}
	members, err := h.repo.List(c, chatmember.Filter{
		ChatID: ch.ID,
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

	if _, err := c.Bot.EditMessageText(
		c,
		botapi.Username("H4venflood"),
		8,
		rolepost.Render(rolepost.BuildRoleStates(members)),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	); err != nil {
		return fmt.Errorf("update roles: %w", err)
	}

	return nil
}

func (h *Handler) RemoveEmoji(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	target, ok := cctx.MustArgs(c).User()

	if !ok || !moderator.Permitted(cctx.MustPermission(c)) {
		target = moderator
	}
	ch := cctx.MustChat(c)
	if err := h.repo.SetEmoji(c, ch.ID, target.ID(), ""); err != nil {
		return fmt.Errorf("remove chat emoji: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.ChatMember.RemoveEmoji.Success, i18n.CmdChatMemberRemoveEmojiSuccessData{
		User: tghtml.MemberMentionCustom(loc, target, false),
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) Ship(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	members, err := h.service.ListHumanPresentChatMembers(c, ch.ID)
	if err != nil {
		return fmt.Errorf("ship: list members: %w", err)
	}

	if len(members) < 2 {
		_, err = c.Reply(loc.T(i18n.Cmd.Ship.None, nil))
		return err
	}

	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	first := members[0]
	second := members[1]

	firstMention := tghtml.MemberLink(loc, ch, first)
	secondMention := tghtml.MemberLink(loc, ch, second)

	var msg i18n.MessageID

	switch {
	case first.User.ID == second.User.ID:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Self1,
			i18n.Cmd.Ship.Self2,
			i18n.Cmd.Ship.Self3,
			i18n.Cmd.Ship.Self4,
			i18n.Cmd.Ship.Self5,
			i18n.Cmd.Ship.Self6,
			i18n.Cmd.Ship.Self7,
			i18n.Cmd.Ship.Self8,
		)

	case first.User.IsBot && second.User.IsBot:
		msg = randomShipMessage(
			i18n.Cmd.Ship.BotBot1,
			i18n.Cmd.Ship.BotBot2,
			i18n.Cmd.Ship.BotBot3,
			i18n.Cmd.Ship.BotBot4,
			i18n.Cmd.Ship.BotBot5,
			i18n.Cmd.Ship.BotBot6,
			i18n.Cmd.Ship.BotBot7,
			i18n.Cmd.Ship.BotBot8,
		)

	case first.User.IsBot || second.User.IsBot:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Bot1,
			i18n.Cmd.Ship.Bot2,
			i18n.Cmd.Ship.Bot3,
			i18n.Cmd.Ship.Bot4,
			i18n.Cmd.Ship.Bot5,
			i18n.Cmd.Ship.Bot6,
			i18n.Cmd.Ship.Bot7,
			i18n.Cmd.Ship.Bot8,
		)

	default:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Normal1,
			i18n.Cmd.Ship.Normal2,
			i18n.Cmd.Ship.Normal3,
			i18n.Cmd.Ship.Normal4,
			i18n.Cmd.Ship.Normal5,
			i18n.Cmd.Ship.Normal6,
			i18n.Cmd.Ship.Normal7,
			i18n.Cmd.Ship.Normal8,
		)
	}

	_, err = c.Reply(
		loc.T(
			msg,
			i18n.CmdShipNormal1Data{
				First:  firstMention,
				Second: secondMention,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func randomShipMessage(ids ...i18n.MessageID) i18n.MessageID {
	return ids[rand.IntN(len(ids))]
}

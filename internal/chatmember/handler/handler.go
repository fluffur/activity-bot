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
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/chatmembers"
	"activity-bot/internal/utils/participant"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

const CategoryChatMember command.Category = "chat_member"

type Handler struct {
	repo    chatmember.Repository
	service *chatmember.Service
}

func NewHandler(repo chatmember.Repository, service *chatmember.Service) *Handler {
	return &Handler{repo: repo, service: service}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"set_emoji_chat",
			h.SetEmoji,
			i18n.Cmd.ChatMember.SetEmoji.Desc,
			CategoryChatMember,
			option.WithAliases("значок", "чат эмоджи"),
			option.IgnorePermissionCheck(),
			option.WithRules(rule.User().Optional(), rule.Text()),
			option.WithPermission(permission.StatusModerator),
		),
		action.NewCommand(
			"show_emoji_chat",
			h.ShowEmoji,
			i18n.Cmd.ChatMember.ShowEmoji.Desc,
			CategoryChatMember,
			option.WithAliases("значок", "чат эмоджи"),
			option.WithRules(rule.User().Optional()),
		),
		action.NewCommand(
			"update",
			h.UpdateChatMembers,
			i18n.Cmd.ChatMember.Update.Desc,
			CategoryChatMember,
			option.WithAliases("обновить чат", "чат обновить"),
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

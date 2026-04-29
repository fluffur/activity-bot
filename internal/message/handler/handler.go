package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/message"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"activity-bot/internal/rp"
	rptemplate "activity-bot/internal/rp/template"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/cohesion-org/deepseek-go"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type Handler struct {
	service        *message.Service
	memberService  *member.Service
	chatService    *chat.Service
	rpService      *rp.Service
	deepseekClient *deepseek.Client
}

func New(service *message.Service, memberService *member.Service, chatService *chat.Service, rpService *rp.Service, deepseekClient *deepseek.Client) *Handler {
	return &Handler{service, memberService, chatService, rpService, deepseekClient}
}

func (h *Handler) Bot(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("bot: get chat: %w", err)
	}
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: "Отвечай коротко, 1-2 предложения. " + c.AISystemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: ctx.TextOrDefault(""),
			},
		},
	}
	resp, err := h.deepseekClient.CreateChatCompletion(ctx.StdContext(), request)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось получить ответ от бота"))
		return fmt.Errorf("bot: create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		_ = ctx.ReplyOnly(u, options.WithText("Бот не вернул ответ"))
		return nil
	}

	return ctx.ReplyOnly(u, options.WithText(resp.Choices[0].Message.Content))
}

func (h *Handler) Message(ctx *command.Context, u *ext.Update) error {
	msg := u.EffectiveMessage
	if msg.Action != nil {
		switch msg.Action.(type) {
		case *tg.MessageActionChatJoinedByLink, *tg.MessageActionChatAddUser, *tg.MessageActionChatDeleteUser:
			return nil
		}
	}

	effectiveSender := u.EffectiveUser()
	effectiveChat := u.EffectiveChat()
	if effectiveSender == nil || effectiveChat == nil {
		return nil
	}
	if _, err := h.memberService.EnsureMemberExists(
		ctx.StdContext(),
		effectiveChat.GetID(),
		effectiveSender.GetID(),
		effectiveSender.Username,
		effectiveSender.FirstName,
		effectiveSender.LastName,
		msg.FromRank,
		effectiveSender.Bot,
	); err != nil {
		return fmt.Errorf("message: ensure member exists: %w", err)
	}

	return h.service.Save(
		ctx.StdContext(),
		effectiveChat.GetID(),
		effectiveSender.GetID(),
		int64(msg.ID),
	)
}

func (h *Handler) HandleRPCommand(ctx *command.Context, u *ext.Update) error {
	msg := u.EffectiveMessage
	effectiveSender := u.EffectiveUser()
	effectiveChat := u.EffectiveChat()
	if msg == nil || effectiveSender == nil || effectiveChat == nil {
		return errors.New("no effective message")
	}

	chatObj, err := ctx.Chat()
	if err != nil {
		return dispatcher.ContinueGroups
	}

	text, ok := rpMatchText(msg.Text, chatObj)
	if !ok {
		return dispatcher.ContinueGroups
	}

	cmd, ok, err := h.rpService.Match(ctx.StdContext(), effectiveChat.GetID(), text)
	if err != nil {
		return err
	}
	if !ok {
		return dispatcher.ContinueGroups
	}

	target, err := ctx.ResolveUser(command.AllowBots, command.UserFromArgs, command.UserFromReply)
	if err != nil {
		return err
	}

	actor, err := h.memberService.GetChatMember(ctx.StdContext(), effectiveChat.GetID(), effectiveSender.GetID())
	if err != nil {
		return dispatcher.ContinueGroups
	}

	detail, speech := extractRPRefinements(text, cmd.Trigger, msg.Text, msg.Entities, target.User.ID, target.User.Username)

	eb := &entity.Builder{}
	renderRPTemplate(eb, cmd, actor, *target, detail, speech)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func rpMatchText(raw string, chat model.Chat) (string, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", false
	}

	prefixes := []string{"фм", "!", "/", "."}
	if chat.CommandPrefix != "" {
		prefixes = append(prefixes, chat.CommandPrefix)
	}

	for _, prefix := range prefixes {
		if hasPrefixIgnoreCase(text, prefix) {
			return strings.TrimSpace(trimPrefixIgnoreCase(text, prefix)), true
		}
	}

	if !chat.AllowPrefixless {
		return "", false
	}

	return text, true
}

func hasPrefixIgnoreCase(s, prefix string) bool {
	sRunes := []rune(s)
	prefixRunes := []rune(prefix)
	if len(prefixRunes) > len(sRunes) {
		return false
	}
	return strings.EqualFold(string(sRunes[:len(prefixRunes)]), prefix)
}

func trimPrefixIgnoreCase(s, prefix string) string {
	sRunes := []rune(s)
	prefixRunes := []rune(prefix)
	if len(prefixRunes) > len(sRunes) {
		return s
	}
	if strings.EqualFold(string(sRunes[:len(prefixRunes)]), prefix) {
		return string(sRunes[len(prefixRunes):])
	}
	return s
}

func (h *Handler) resolveRPCommandTarget(ctx *command.Context, u *ext.Update) (*model.ChatMember, error) {
	msg := u.EffectiveMessage
	if msg == nil {
		return nil, nil
	}
	chatID := u.EffectiveChat().GetID()

	if replyTo, ok := msg.GetReplyTo(); ok {
		if header, ok := replyTo.(*tg.MessageReplyHeader); ok {
			replyMessages, err := ctx.GetMessages(chatID, []tg.InputMessageClass{&tg.InputMessageID{ID: header.ReplyToMsgID}})
			if err == nil && len(replyMessages) > 0 {
				if targetID, found := messageUserID(replyMessages[0]); found {
					m, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, targetID)
					if err == nil {
						return &m, nil
					}
				}
			}
		}
	}

	for _, e := range msg.Entities {
		switch ent := e.(type) {
		case *tg.MessageEntityMentionName:
			m, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, ent.UserID)
			if err == nil {
				return &m, nil
			}
		case *tg.MessageEntityMention:
			username := parseMentionUsername(msg.Text, ent)
			if username == "" {
				continue
			}
			m, err := h.memberService.GetChatMemberByUsername(ctx.StdContext(), chatID, username)
			if err == nil {
				return &m, nil
			}
		}
	}

	return nil, nil
}

func messageUserID(msg tg.MessageClass) (int64, bool) {
	switch cast := msg.(type) {
	case *tg.Message:
		if fromID, ok := cast.GetFromID(); ok {
			if peerUser, ok := fromID.(*tg.PeerUser); ok {
				return peerUser.UserID, true
			}
		}
	}
	return 0, false
}

func parseMentionUsername(text string, mention *tg.MessageEntityMention) string {
	encoded := utf16.Encode([]rune(text))
	start := mention.Offset
	end := start + mention.Length
	if start < 0 || end > len(encoded) || start >= end {
		return ""
	}

	segment := string(utf16.Decode(encoded[start:end]))
	return strings.TrimPrefix(segment, "@")
}

func renderRPTemplate(eb *entity.Builder, cmd model.RPCommand, actor model.ChatMember, target model.ChatMember, detail, speech string) {

	if len(cmd.Emoji) > 0 {
		helpers.DisplayEmoji(eb, cmd.Emoji)
		eb.Plain("| ")
	}
	template := rptemplate.Normalize(strings.TrimSpace(cmd.Template))
	if template == "" {
		helpers.WriteRoleEmojiMention(eb, actor)
		eb.Plain(" ")
		if len(cmd.Emoji) > 0 {
			helpers.DisplayEmoji(eb, cmd.Emoji)
			eb.Plain(" ")
		}
		eb.Plain("использует действие «")
		eb.Plain(cmd.Trigger)
		eb.Plain("» на ")
		helpers.WriteRoleEmojiMention(eb, target)
		return
	}

	replaced := strings.ReplaceAll(template, "{command}", cmd.Trigger)
	replaced = strings.ReplaceAll(replaced, "{detail}", detail)
	replaced = rptemplate.ResolveVariants(replaced, actor.User.Gender, target.User.Gender)
	parts := strings.Split(replaced, "{actor}")
	for i, part := range parts {
		targetParts := strings.Split(part, "{target}")
		for j, chunk := range targetParts {
			if chunk != "" {
				eb.Plain(chunk)
			}
			if j < len(targetParts)-1 {
				helpers.WriteMemberMention(eb, target)
			}
		}
		if i < len(parts)-1 {
			helpers.WriteMemberMention(eb, actor)
		}
	}

	if detail != "" && !strings.Contains(template, "{detail}") {
		eb.Plain(" (")
		eb.Plain(detail)
		eb.Plain(")")
	}
	_ = speech
}

func extractRPRefinements(
	text string,
	trigger string,
	fullText string,
	entities []tg.MessageEntityClass,
	targetID int64,
	targetUsername string,
) (string, string) {
	rest, ok := cutRPTriggerPrefix(text, trigger)
	if !ok {
		return "", ""
	}
	if rest == "" {
		return "", ""
	}
	rest = stripTargetMentionFromRest(rest, fullText, entities, targetID, targetUsername)
	if rest == "" {
		return "", ""
	}

	for len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t') {
		rest = rest[1:]
	}
	if rest == "" {
		return "", ""
	}

	detail := strings.Join(strings.Fields(rest), " ")
	return detail, ""
}

func stripTargetMentionFromRest(rest, fullText string, entities []tg.MessageEntityClass, targetID int64, targetUsername string) string {
	result := rest
	trimmedUsername := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(targetUsername)), "@")

	for _, e := range entities {
		switch ent := e.(type) {
		case *tg.MessageEntityMentionName:
			if ent.UserID != targetID {
				continue
			}
			mention := rpExtractEntity(fullText, ent)
			if mention != "" {
				result = strings.Replace(result, mention, "", 1)
			}
		case *tg.MessageEntityMention:
			mention := rpExtractEntity(fullText, ent)
			if mention == "" {
				continue
			}
			username := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(mention)), "@")
			if trimmedUsername != "" && username == trimmedUsername {
				result = strings.Replace(result, mention, "", 1)
			}
		}
	}

	if trimmedUsername != "" {
		result = strings.Replace(result, "@"+trimmedUsername, "", 1)
		result = strings.Replace(result, "@"+strings.ToUpper(trimmedUsername), "", 1)
	}

	return strings.TrimSpace(result)
}

func rpExtractEntity(text string, e tg.MessageEntityClass) string {
	encoded := utf16.Encode([]rune(text))
	start := e.GetOffset()
	end := start + e.GetLength()
	if start < 0 || end > len(encoded) || start >= end {
		return ""
	}
	return string(utf16.Decode(encoded[start:end]))
}

func cutRPTriggerPrefix(text, trigger string) (string, bool) {
	textRunes := []rune(text)
	triggerRunes := []rune(strings.TrimSpace(trigger))
	i, j := 0, 0

	for j < len(triggerRunes) {
		if triggerRunes[j] == ' ' || triggerRunes[j] == '\t' || triggerRunes[j] == '\n' || triggerRunes[j] == '\r' {
			if i >= len(textRunes) || !(textRunes[i] == ' ' || textRunes[i] == '\t' || textRunes[i] == '\n' || textRunes[i] == '\r') {
				return "", false
			}
			for j < len(triggerRunes) && (triggerRunes[j] == ' ' || triggerRunes[j] == '\t' || triggerRunes[j] == '\n' || triggerRunes[j] == '\r') {
				j++
			}
			for i < len(textRunes) && (textRunes[i] == ' ' || textRunes[i] == '\t' || textRunes[i] == '\n' || textRunes[i] == '\r') {
				i++
			}
			continue
		}
		if i >= len(textRunes) || !strings.EqualFold(string(textRunes[i]), string(triggerRunes[j])) {
			return "", false
		}
		i++
		j++
	}

	return string(textRunes[i:]), true
}

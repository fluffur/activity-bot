package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	rptemplate "activity-bot/internal/rp/template"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

func (h *Handler) SetRPCommand(ctx *command.Context, u *ext.Update) error {
	chatObj, err := ctx.Chat()
	if err != nil {
		return err
	}

	trigger, template, emojis, err := parseRPCreateArgs(ctx.RawArgs, ctx.RawArgsEntities)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("❌ "+err.Error()))
	}

	sender := u.EffectiveUser()
	if sender == nil {
		return nil
	}

	if err := h.rpService.Upsert(ctx.StdContext(), model.RPCommand{
		ChatID:    chatObj.ID,
		Trigger:   trigger,
		Template:  template,
		Emoji:     emojis,
		CreatedBy: sender.GetID(),
	}); err != nil {
		return err
	}

	eb := &entity.Builder{}
	eb.Plain("✅ РП-команда ")
	eb.Code(trigger)
	eb.Plain(" сохранена. Вызывается сообщением: ")
	eb.Code(trigger)
	eb.Plain(" + упоминание/реплай на пользователя.")
	if template != "" {
		eb.Plain("\nШаблон: ")
		eb.Code(template)
	}
	if len(emojis) > 0 {
		eb.Plain("\nЭмодзи: ")
		helpers.DisplayEmoji(eb, emojis)
	}

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) PreviewRPTemplate(ctx *command.Context, u *ext.Update) error {
	template := strings.TrimSpace(ctx.RawArgs)
	if template == "" {
		return ctx.ReplyOnly(u, options.WithText("❌ Укажите шаблон. Пример: рп превью {A} обнял(а) {B}"))
	}

	template = rptemplate.Normalize(template)

	eb := &entity.Builder{}
	eb.Bold("Превью РП-шаблона")
	eb.Plain("\n\n")
	eb.Code(template)
	eb.Plain("\n\n")

	render := func(actorName, actorGender, targetName, targetGender string) string {
		line := rptemplate.ResolveVariants(template, actorGender, targetGender)
		line = strings.ReplaceAll(line, "{command}", "обнять")
		line = strings.ReplaceAll(line, "{actor}", actorName)
		line = strings.ReplaceAll(line, "{target}", targetName)
		return line
	}

	eb.Plain("♂♂: ")
	eb.Plain(render("Алексей", model.GenderMale, "Дмитрий", model.GenderMale))
	eb.Plain("\n♂♀: ")
	eb.Plain(render("Алексей", model.GenderMale, "Анна", model.GenderFemale))
	eb.Plain("\n♀♂: ")
	eb.Plain(render("Анна", model.GenderFemale, "Дмитрий", model.GenderMale))
	eb.Plain("\n♀♀: ")
	eb.Plain(render("Анна", model.GenderFemale, "Мария", model.GenderFemale))

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) RemoveRPCommand(ctx *command.Context, u *ext.Update) error {
	chatObj, err := ctx.Chat()
	if err != nil {
		return err
	}

	trigger := strings.TrimSpace(ctx.RawArgs)
	if trigger == "" {
		return ctx.ReplyOnly(u, options.WithText("❌ Укажите триггер, например: -рп обнять"))
	}

	if err := h.rpService.Delete(ctx.StdContext(), chatObj.ID, trigger); err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("🗑 РП-команда '%s' удалена", trigger)))
}

func (h *Handler) ListRPCommands(ctx *command.Context, u *ext.Update) error {
	chatObj, err := ctx.Chat()
	if err != nil {
		return err
	}

	commands, err := h.rpService.ListByChat(ctx.StdContext(), chatObj.ID)
	if err != nil {
		return err
	}

	if len(commands) == 0 {
		eb := &entity.Builder{}
		eb.Plain("РП-команд пока нет. Добавить: ")
		eb.Code("+рп обнять 🤗 {A} обнял(а) {B}")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	eb := &entity.Builder{}
	eb.Bold("РП-команды чата")
	eb.Plain("\n\n")
	for _, cmd := range commands {
		eb.Plain("• ")
		eb.Code(cmd.Trigger)
		if len(cmd.Emoji) > 0 {
			eb.Plain(" ")
			helpers.DisplayEmoji(eb, cmd.Emoji)
		}
		if cmd.Template != "" {
			eb.Plain(" — ")
			eb.Code(cmd.Template)
		}
		eb.Plain("\n")
	}
	eb.Plain("\nЭлементы шаблона:\n")
	token := eb.Token()
	eb.Plain("• ")
	eb.Code("{actor}")
	eb.Plain(" — упоминание автора действия\n")
	eb.Plain("• ")
	eb.Code("{target}")
	eb.Plain(" — упоминание цели действия\n")
	eb.Plain("• ")
	eb.Code("{command}")
	eb.Plain(" — текст триггера команды\n")
	eb.Plain("• ")
	eb.Code("{A}")
	eb.Plain(" / ")
	eb.Code("{B}")
	eb.Plain(" — короткие аналоги для {actor}/{target}\n")
	eb.Plain("• ")
	eb.Code("обнял(а)")
	eb.Plain(" — короткая форма для варианта {actor:обнял|обняла}\n")
	eb.Plain("• ")
	eb.Code("{actor:обнял|обняла}")
	eb.Plain(" — вариант по полу автора\n")
	eb.Plain("• ")
	eb.Code("{target:обнял|обняла}")
	eb.Plain(" — вариант по полу цели\n")
	eb.Plain("• ")
	eb.Code("{pair:mm|mf|fm|ff}")
	eb.Plain(" — вариант по сочетанию полов (автор/цель)")
	token.Apply(eb, entity.Blockquote(true))

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func parseRPCreateArgs(raw string, entities []tg.MessageEntityClass) (string, string, model.Emojis, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", nil, fmt.Errorf("пример: +рп обнять 🤗 {A} обнял(а) {B}")
	}

	tokens := strings.Fields(raw)
	for i, token := range tokens {
		tokenEntities := entitiesInToken(raw, entities, token)
		emojis := helpers.ExtractEmoji(token, tokenEntities)
		if len(emojis) == 0 {
			continue
		}

		trigger := strings.TrimSpace(strings.Join(tokens[:i], " "))
		if trigger == "" {
			return "", "", nil, fmt.Errorf("укажите триггер перед эмодзи")
		}

		template := strings.TrimSpace(strings.Join(tokens[i+1:], " "))
		return trigger, rptemplate.Normalize(template), emojis, nil
	}

	trigger, template := rptemplate.SplitTriggerAndTemplate(raw)
	if trigger == "" {
		return "", "", nil, fmt.Errorf("не удалось определить триггер, пример: +рп обнять 🤗 {A} обнял(а) {B}")
	}
	return trigger, rptemplate.Normalize(template), nil, nil
}

func entitiesInToken(fullRaw string, entities []tg.MessageEntityClass, token string) []tg.MessageEntityClass {
	tokenStartBytes := strings.Index(fullRaw, token)
	if tokenStartBytes < 0 {
		return nil
	}
	tokenEndBytes := tokenStartBytes + len(token)

	tokenStartUTF16 := len(utf16.Encode([]rune(fullRaw[:tokenStartBytes])))
	tokenLenUTF16 := len(utf16.Encode([]rune(fullRaw[tokenStartBytes:tokenEndBytes])))
	tokenEndUTF16 := tokenStartUTF16 + tokenLenUTF16

	var result []tg.MessageEntityClass
	for _, e := range entities {
		start := e.GetOffset()
		end := start + e.GetLength()
		if start >= tokenStartUTF16 && end <= tokenEndUTF16 {
			switch cast := e.(type) {
			case *tg.MessageEntityCustomEmoji:
				result = append(result, &tg.MessageEntityCustomEmoji{
					Offset:     cast.Offset - tokenStartUTF16,
					Length:     cast.Length,
					DocumentID: cast.DocumentID,
				})
			}
		}
	}

	return result
}

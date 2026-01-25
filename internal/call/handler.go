package call

import (
	"activity-bot/internal/admin"
	"regexp"
)

const mentionsPerMessage = 5

var emojis = []string{
	"🔔", "👋", "⚡", "📢", "🔥",
	"🟡", "🟢", "🔹", "🔸", "⭐",
	"✨", "🏆", "🎯", "🚨", "🛎️",
	"🥁", "📣", "🎉", "🟠", "🟣",
}

type Handler struct {
	adminService *admin.Service
	callRe       *regexp.Regexp
}

func NewHandler(adminService *admin.Service, callRe *regexp.Regexp) *Handler {
	return &Handler{adminService, callRe}
}

//
//func (h *Handler) Call(ctx context.Context, b *bot.Bot, update *models.Update) {
//	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.Message.Chat.ID, update.Message.From.ID) {
//		helpers.SendMessage(ctx, b, update, "Команда доступна только администраторам бота и создателю чата")
//		return
//	}
//	if update.Message == nil {
//		return
//	}
//
//	matches := h.callRe.FindStringSubmatch(update.Message.Text)
//
//	message := ""
//	if len(matches) > 1 {
//		message = strings.TrimSpace(matches[1])
//	}
//
//	administrators, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
//		ChatID: update.Message.Chat.ID,
//	})
//	if err != nil {
//		log.Println("GetChatAdministrators", err)
//		helpers.SendMessage(ctx, b, update, "Не удалось созвать пользователей")
//		return
//	}
//
//	var users []*models.User
//	for _, a := range administrators {
//		var user *models.User
//		switch a.Type {
//		case models.ChatMemberTypeAdministrator:
//			user = &a.Administrator.User
//		case models.ChatMemberTypeOwner:
//			user = a.Owner.User
//		default:
//			continue
//		}
//		if user.IsBot {
//			continue
//		}
//		users = append(users, user)
//	}
//
//	for i := 0; i < len(users); i += mentionsPerMessage {
//		end := i + mentionsPerMessage
//		if end > len(users) {
//			end = len(users)
//		}
//
//		var sb strings.Builder
//
//		if i == 0 && message != "" {
//			sb.WriteString(fmt.Sprintf("%s\n\n", message))
//		}
//
//		for _, user := range users[i:end] {
//			emoji := emojis[rand.Intn(len(emojis))]
//			sb.WriteString(fmt.Sprintf("%s ", helpers.Mention(helpers.MapUser(user), emoji)))
//		}
//
//		helpers.SendMessage(ctx, b, update, sb.String())
//	}
//}

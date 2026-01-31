package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/model"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"log"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	statsService *stats.Service
	userService  *user.Service
}

func NewHandler(statsService *stats.Service, userService *user.Service) *Handler {
	return &Handler{statsService, userService}
}

var countRe = regexp.MustCompile(`—\s*(\d+)`)

func (h *Handler) Import(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	msg := ctx.EffectiveMessage.ReplyToMessage
	if msg == nil {
		return nil
	}

	msgText := msg.GetText()
	log.Println(msgText)

	var period stats.ReportPeriod
	switch {
	case strings.Contains(msgText, "за сутки"):
		period = stats.PeriodDay
	case strings.Contains(msgText, "за неделю"):
		period = stats.PeriodSevenDays
	case strings.Contains(msgText, "за месяц"):
		period = stats.PeriodThirtyDays
	case strings.Contains(msgText, "за всё время"):
		period = stats.PeriodAll
	default:
		slog.Warn("Could not detect statistics period")
		return nil
	}

	log.Printf("Detected period: %s", period)

	var userIDs []int64
	var counts []int32

	for _, e := range msg.Entities {
		var u model.User
		var err error

		switch e.Type {
		case "text_link":
			if !strings.HasPrefix(e.Url, "https://t.me/") {
				continue
			}
			username := strings.TrimPrefix(e.Url, "https://t.me/")
			u, err = h.userService.GetUserByUsername(username)
		case "mention":
			username := msg.Text[e.Offset+1 : e.Offset+e.Length]
			u, err = h.userService.GetUserByUsername(username)
		case "text_mention":
			if e.User != nil {
				u, err = h.userService.EnsureUserExists(e.User.Id, e.User.Username, e.User.FirstName, e.User.LastName)
			}
		default:
			continue
		}

		if err != nil {
			slog.Warn("User not found or error", "type", e.Type, "error", err)
			continue
		}

		textRunes := []rune(msg.Text)
		startSearch := e.Offset + e.Length
		endLine := len(textRunes)
		for i := int(startSearch); i < len(textRunes); i++ {
			if textRunes[i] == '\n' {
				endLine = i
				break
			}
		}

		textAfter := string(textRunes[startSearch:endLine])
		matches := countRe.FindStringSubmatch(textAfter)
		if len(matches) < 2 {
			continue
		}

		count, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("cannot parse message count for %s: %v", *u.Username, err)
			continue
		}

		log.Printf("Username: %s, Messages: %d\n", *u.Username, count)

		userIDs = append(userIDs, u.ID)
		counts = append(counts, int32(count))
	}

	if err := h.statsService.ImportActivity(ctx.EffectiveChat.Id, period, userIDs, counts); err != nil {
		slog.Error("Failed to import activity", err.Error())
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Импорт произведён успешно", nil)
	return err
}

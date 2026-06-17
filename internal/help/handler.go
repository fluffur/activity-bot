package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"log"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot        *botapi.Bot
	translator *i18n.Service
}

func NewHandler(bot *botapi.Bot, translator *i18n.Service) *Handler {
	return &Handler{bot, translator}
}

func (h *Handler) Register() {
	h.bot.OnCommand("help", "Помощь", h.Help, func(c *botapi.Context) bool {
		m, err := cctx.ChatMember(c.Context)
		if err != nil {
			log.Printf("%s", err.Error())

			return false
		}

		log.Printf("%+v", m)

		return true
	})
}

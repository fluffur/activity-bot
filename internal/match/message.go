package match

import (
	"github.com/go-telegram/bot/models"
)

func Message(update *models.Update) bool {
	return update.Message != nil
}

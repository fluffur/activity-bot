package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"

	"github.com/gotd/td/telegram/message/entity"
)

func FormatInactiveMembers(members []model.InactiveMember) string {
	eb := &entity.Builder{}
	WriteInactiveMembers(eb, members)
	res, _ := eb.Complete()
	return res
}

func WriteInactiveMembers(eb *entity.Builder, members []model.InactiveMember) {
	eb.Bold("😴 Неактивные участники (более 1 суток)\n\n")

	for i, m := range members {
		userTitle := m.Member.Tag
		if userTitle == "" {
			userTitle = m.Member.User.FirstName
		}
		eb.Plain(fmt.Sprintf("%d. ", i+1))
		helpers.WriteRoleEmojiLink(eb, m.Member)
		eb.Plain(" — ")

		if !m.LastActivity.IsZero() {
			eb.FormattedDate("date", true, false, false, false, false, false, int(m.LastActivity.Unix()))
		} else {
			eb.Plain("не писал ни разу")
		}

		eb.Plain("\n")
	}
}

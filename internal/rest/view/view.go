package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

func WriteRestSet(eb *entity.Builder, member model.ChatMember, date time.Time, reason string) {
	if reason != "" {
		eb.Plain(reason)
		eb.Plain("\n\n")
	}
	eb.Plain("Участник ")
	helpers.WriteRoleEmojiLink(eb, member)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(member.User.Gender, "добавлен", "добавлена"))
	eb.Plain(" в рест до ")
	helpers.FormattedDate(eb, date)
}

func WriteRestRequest(eb *entity.Builder, user model.ChatMember, date time.Time, reason string) {
	if reason != "" {
		eb.Plain(reason)
		eb.Plain("\n\n")
	}
	eb.Plain("Для участника ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" запрошен рест до ")
	helpers.FormattedDate(eb, date)
}

func WriteRestShow(eb *entity.Builder, m model.ChatMember) {
	if m.RestUntil.IsZero() {
		eb.Plain("Участник ")
		helpers.WriteRoleEmojiLink(eb, m)
		eb.Plain(" не находится в ресте")
		return
	}

	message := " находится в ресте до "
	if m.RestUntil.Before(time.Now()) {
		eb.Plain("Рест ")
		helpers.WriteRoleEmojiLink(eb, m)
		eb.Plain(" был завершен ")
		helpers.FormattedDate(eb, m.RestUntil)
	} else {
		helpers.WriteRoleEmojiLink(eb, m)
		eb.Plain(message)
		helpers.FormattedDate(eb, m.RestUntil)
	}

	if m.RestReason != "" {
		eb.Plain("\n\nПричина: ")
		eb.Plain(m.RestReason) // Reason in DB is usually plain text or HTML, for now assuming it might benefit from builder
	}
}

func WriteRestEnded(eb *entity.Builder, user model.ChatMember, isSelf bool) {
	if isSelf {
		eb.Plain("Вы успешно удалены из реста")
		return
	}
	eb.Plain("Участник ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" успешно ")
	eb.Plain(helpers.Gendered(user.User.Gender, "удалён", "удалена"))
	eb.Plain(" из реста")
}

func WriteRestNotInRest(eb *entity.Builder, user model.ChatMember, isSelf bool) {
	if isSelf {
		eb.Plain("Вы не находитесь в ресте")
		return
	}
	eb.Plain("Пользователь ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" не находится в ресте")
}

func WriteRestRequestApproved(eb *entity.Builder, user model.ChatMember, restUntil time.Time) {
	eb.Plain("Запрос одобрен. У ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" рест до ")
	helpers.FormattedDate(eb, restUntil)
}

func WriteRestRequestRejected(eb *entity.Builder, user *model.ChatMember) {
	if user == nil {
		eb.Plain("Запрос на рест отклонён")
		return
	}
	eb.Plain("Запрос на рест для ")
	helpers.WriteRoleEmojiLink(eb, *user)
	eb.Plain(" отклонён")
}

func WriteRestRequests(eb *entity.Builder, requests []model.ApprovedRestRequest) {
	if len(requests) == 0 {
		eb.Plain("Список рестов пуст")
		return
	}

	var cm model.ChatMember
	if len(requests) > 0 {
		cm = requests[0].ChatMember
	}

	eb.Plain("Список рестов ")
	helpers.WriteRoleEmojiLink(eb, cm)
	eb.Plain(":\n")

	type group struct {
		title string
		reqs  []model.ApprovedRestRequest
	}
	groups := []group{
		{"Одобренные:", nil},
		{"Отклонённые:", nil},
	}

	for _, r := range requests {
		if r.Status == "approved" {
			groups[0].reqs = append(groups[0].reqs, r)
		} else if r.Status == "rejected" {
			groups[1].reqs = append(groups[1].reqs, r)
		}
	}

	for _, g := range groups {
		if len(g.reqs) == 0 {
			continue
		}

		eb.Plain("\n")
		eb.Plain(g.title)
		eb.Plain("\n")

		token := eb.Token()
		for i, r := range g.reqs {
			eb.Code(fmt.Sprintf("%d", i+1))
			eb.Plain(" ")
			writeRestActiveMessageEB(eb, r)
			eb.Plain(" Срок окончания ")
			helpers.FormattedDate(eb, r.RestUntil)
			if r.Reason != "" {
				eb.Plain(fmt.Sprintf(" (%s)", r.Reason))
			}
			eb.Plain("\n• Запрошено ")
			helpers.FormattedDate(eb, r.RequestedAt)
			if !r.UpdatedAt.IsZero() {
				eb.Plain("\n• Одобрено ")
				helpers.FormattedDate(eb, r.UpdatedAt)
			}
			if i < len(g.reqs)-1 {
				eb.Plain("\n\n")
			}
		}
		token.Apply(eb, entity.Blockquote(true))
	}

	eb.Plain("\nЧтобы удалить определенный рест введите команду ")
	eb.Code("удалить рест @участник номер")
}

func writeRestActiveMessageEB(eb *entity.Builder, rr model.ApprovedRestRequest) {
	now := time.Now()
	if !rr.ChatMember.IsRestActive(now) || !rr.ChatMember.RestUntil.Equal(rr.RestUntil) {
		helpers.WriteDangerEmoji(eb)
		eb.Plain(" Недействителен\n")
	} else {
		helpers.WriteSuccessEmoji(eb)
		eb.Plain(" Действителен\n")
	}
}

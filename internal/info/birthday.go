package info

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/utils"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"sort"
	"strings"
)

type BirthdayMember struct {
	Name string
	Day  int
}

type BirthdayMonth struct {
	Number  int
	Name    string
	Members []BirthdayMember
}

var birthdayMonths = [...]string{
	"январь",
	"февраль",
	"март",
	"апрель",
	"май",
	"июнь",
	"июль",
	"август",
	"сентябрь",
	"октябрь",
	"ноябрь",
	"декабрь",
}

func BuildBirthdayMonths(members []chatmember.ChatMember) []BirthdayMonth {
	months := make([]BirthdayMonth, 12)

	for i := range months {
		months[i].Name = tghtml.Blockquote(utils.UcFirst(birthdayMonths[i]))
		months[i].Number = i + 1
	}

	for _, member := range members {
		if member.IsLeft() || member.Birthday.IsZero() {
			continue
		}

		birthday := member.Birthday

		months[birthday.Month()-1].Members = append(
			months[birthday.Month()-1].Members,
			BirthdayMember{
				Name: member.Display("", false),
				Day:  birthday.Day(),
			},
		)
	}

	for i := range months {
		sort.Slice(
			months[i].Members,
			func(a, b int) bool {
				return months[i].Members[a].Day < months[i].Members[b].Day
			},
		)
	}

	result := make([]BirthdayMonth, 0, 12)

	for _, month := range months {
		if len(month.Members) > 0 {
			result = append(result, month)
		}
	}

	return result
}

func RenderBirthdays(members []chatmember.ChatMember) string {
	months := BuildBirthdayMonths(members)

	var b strings.Builder

	b.WriteString("Дни рождения участников\n\n")

	for _, month := range months {
		b.WriteString(month.Name)
		b.WriteString("\n")

		for _, member := range month.Members {
			b.WriteString(member.Name)
			b.WriteString(" — ")
			b.WriteString(fmt.Sprintf("%02d.%02d", member.Day, month.Number))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	return b.String()
}

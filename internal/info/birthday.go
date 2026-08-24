package info

import (
	"activity-bot/internal/chatmember"
	"sort"
	"strconv"
	"strings"
)

type BirthdayMember struct {
	Name string
	Day  int
}

type BirthdayMonth struct {
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
		months[i].Name = birthdayMonths[i]
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
				return strings.ToLower(months[i].Members[a].Name) <
					strings.ToLower(months[i].Members[b].Name)
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
			b.WriteString(strconv.Itoa(member.Day))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	return b.String()
}

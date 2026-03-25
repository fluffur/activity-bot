package helpers

import "activity-bot/internal/model"

func Gendered(gender, male, female string) string {
	switch gender {
	case model.GenderMale:
		return male
	case model.GenderFemale:
		return female
	default:
		return male
	}
}

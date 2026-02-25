package helpers

import "activity-bot/internal/model"

// Gendered returns male/female/neutral form based on gender string.
func Gendered(gender, male, female, neutral string) string {
	switch gender {
	case model.GenderMale:
		return male
	case model.GenderFemale:
		return female
	default:
		return neutral
	}
}

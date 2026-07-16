package norm

import (
	"activity-bot/internal/i18n"
	"strings"
	"unicode"
	"unicode/utf8"
)

func IsValidNormName(name string) bool {
	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	length := utf8.RuneCountInString(name)

	if length > 20 || length < 2 {
		return false
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}

	return true
}

func LocalisedNormName(loc *i18n.Localizer, name string) string {
	switch name {
	case GeneralNormName:
		return loc.T(i18n.Cmd.AddNorm.NormGeneral, nil)
	default:
		return name
	}
}

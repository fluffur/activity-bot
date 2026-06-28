package norm

import (
	"activity-bot/internal/i18n"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isValidNormName(name string) bool {
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

func LocalisedNormName(t *i18n.Translator, lang string, name string) string {
	switch name {
	case GeneralNormName:
		return t.T(lang, i18n.Cmd.AddNorm.NormGeneral)
	default:
		return name
	}
}

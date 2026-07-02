package i18n

import (
	"activity-bot/internal/user"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizer struct {
	localizer *goi18n.Localizer
}

func (l *Localizer) T(id MessageID, data any) string {
	msg, err := l.localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    string(id),
		TemplateData: data,
	})
	if err != nil {
		return string(id)
	}

	return msg
}

type GenderMessage struct {
	Female  MessageID
	Male    MessageID
	Unknown MessageID
}

func (l *Localizer) TGender(
	gender user.Gender,
	id GenderMessage,
	data any,
) string {
	base := string(id.Male)

	switch {
	case strings.HasSuffix(base, ".male"):
		base = strings.TrimSuffix(base, ".male")
	case strings.HasSuffix(base, ".female"):
		base = strings.TrimSuffix(base, ".female")
	default:
		return l.T(id.Male, data)
	}

	switch gender {
	case user.GenderFemale:
		return l.T(MessageID(base+".female"), data)
	case user.GenderUnknown:
		return l.T(MessageID(base+".unknown"), data)
	default:
		return l.T(MessageID(base+".male"), data)
	}
}

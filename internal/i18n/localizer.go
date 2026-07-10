package i18n

import (
	"activity-bot/internal/user"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizer struct {
	localizer *goi18n.Localizer
}

func (l *Localizer) T(id MessageID, data any, opts ...LocalizeOption) string {
	cfg := &goi18n.LocalizeConfig{
		MessageID:    string(id),
		TemplateData: data,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	msg, err := l.localizer.Localize(cfg)
	if err != nil {
		return string(id)
	}

	return msg
}

type LocalizeOption func(*goi18n.LocalizeConfig)

func WithPluralCount(count any) LocalizeOption {
	return func(cfg *goi18n.LocalizeConfig) {
		cfg.PluralCount = count
	}
}

func WithGender(gender user.Gender) LocalizeOption {
	return func(cfg *goi18n.LocalizeConfig) {
		switch gender {
		case user.GenderFemale:
			cfg.MessageID = cfg.MessageID + ".female"
		case user.GenderUnknown:
			cfg.MessageID = cfg.MessageID + ".unknown"
		default:
			cfg.MessageID = cfg.MessageID + ".male"
		}
	}
}

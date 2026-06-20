package i18n

import (
	"embed"
	"io/fs"
	"path/filepath"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Translator struct {
	bundle *goi18n.Bundle
}

//go:embed locales/*.toml
var localeFS embed.FS

func New() (*Translator, error) {
	bundle := goi18n.NewBundle(language.Russian)

	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	if err := loadLocales(bundle); err != nil {
		return nil, err
	}

	return &Translator{
		bundle: bundle,
	}, nil
}

func loadLocales(bundle *goi18n.Bundle) error {
	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join("locales", entry.Name())

		if _, err := bundle.LoadMessageFileFS(localeFS, path); err != nil {
			return err
		}
	}

	return nil
}

func (s *Translator) T(lang string, messageID MessageID) string {
	return s.TData(lang, messageID, nil)
}

func (s *Translator) TData(lang string, messageID MessageID, data map[string]any) string {
	localizer := goi18n.NewLocalizer(s.bundle, lang)

	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    string(messageID),
		TemplateData: data,
	})
	if err != nil {
		return string(messageID)
	}

	return msg
}

func (s *Translator) TIf(
	lang string,
	condition bool,
	trueKey MessageID,
	falseKey MessageID,
	trueArgs map[string]any,
	falseArgs map[string]any,
) string {
	if condition {
		return s.TData(lang, trueKey, trueArgs)
	}

	return s.TData(lang, falseKey, falseArgs)
}

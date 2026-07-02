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

//go:embed locales/**
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

func (t *Translator) Localizer(lang string) *Localizer {
	return &Localizer{
		localizer: goi18n.NewLocalizer(t.bundle, lang),
	}
}

func (t *Translator) Default() *Localizer {
	return t.Localizer(language.Russian.String())
}

func loadLocales(bundle *goi18n.Bundle) error {
	return fs.WalkDir(localeFS, "locales", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".toml" {
			return nil
		}

		_, err = bundle.LoadMessageFileFS(localeFS, path)
		return err
	})
}

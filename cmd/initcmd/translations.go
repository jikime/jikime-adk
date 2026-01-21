package initcmd

import (
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"jikime-adk-v2/translations"
)

var (
	bundle         *i18n.Bundle
	translationsFS = translations.Files
)

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	loadTranslation("en.yaml")
	loadTranslation("ko.yaml")
	loadTranslation("ja.yaml")
	loadTranslation("zh.yaml")
}

func loadTranslation(filename string) {
	data, err := translationsFS.ReadFile(filename)
	if err != nil {
		log.Printf("translation read error %s: %v", filename, err)
		return
	}
	if _, err := bundle.ParseMessageFileBytes(data, filename); err != nil {
		log.Printf("translation parse error %s: %v", filename, err)
	}
}

func getTranslation(locale string) *i18n.Localizer {
	if locale == "" {
		locale = "en"
	}
	return i18n.NewLocalizer(bundle, locale)
}

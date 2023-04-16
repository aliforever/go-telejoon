package telejoon

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type LanguageConfig struct {
	languages *Languages
	repo      UserLanguageRepository

	changeLanguageState string
}

func NewLanguageConfig(languages *Languages, repo UserLanguageRepository) *LanguageConfig {
	return &LanguageConfig{
		languages: languages,
		repo:      repo,
	}
}

// WithChangeLanguageMenu sets the change language menu state.
func (l *LanguageConfig) WithChangeLanguageMenu(state string) *LanguageConfig {
	l.changeLanguageState = state
	return l
}

type Language struct {
	tag       string
	localizer *i18n.Localizer
}

// Get returns the localized string for the given message ID.
func (l *Language) Get(id string) (string, error) {
	return l.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: id,
	})
}

// MustGet returns the localized string for the given message ID.
// If the message ID is not found, it will panic.
func (l *Language) MustGet(id string, args ...interface{}) string {
	return l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: args,
	})
}

// GetWithParams returns the localized string for the given message ID and parameters.
func (l *Language) GetWithParams(id string, params map[string]interface{}) (string, error) {
	return l.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: params,
	})
}

// MustGetWithParams returns the localized string for the given message ID and parameters.
func (l *Language) MustGetWithParams(id string, params map[string]interface{}) string {
	return l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: params,
	})
}

type Languages struct {
	localizers []Language
}

// GetByTag returns the localizer by the given tag.
func (l *Languages) GetByTag(tag language.Tag) Language {
	for _, localizer := range l.localizers {
		if localizer.tag == tag.String() {
			return localizer
		}
	}

	return Language{}
}

type LanguagesBuilder struct {
	defaultBundle    *i18n.Bundle
	unmarshalFuncs   map[string]i18n.UnmarshalFunc
	messageFilePaths []string
}

func NewLanguageBuilder(defaultBundle language.Tag) *LanguagesBuilder {
	return &LanguagesBuilder{
		defaultBundle: i18n.NewBundle(defaultBundle),
	}
}

// RegisterTomlFormat registers the TOML format for the bundle.
func (lb *LanguagesBuilder) RegisterTomlFormat(tomlFilePaths []string) *LanguagesBuilder {
	lb.defaultBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	lb.messageFilePaths = append(lb.messageFilePaths, tomlFilePaths...)

	return lb
}

// Build builds the bundle.
func (lb *LanguagesBuilder) Build() (*Languages, error) {
	localizers := []Language{}

	for _, path := range lb.messageFilePaths {
		if msgFile, err := lb.defaultBundle.LoadMessageFile(path); err != nil {
			return nil, err
		} else {
			localizers = append(localizers, Language{
				tag:       msgFile.Tag.String(),
				localizer: i18n.NewLocalizer(lb.defaultBundle, msgFile.Tag.String()),
			})
		}
	}

	return &Languages{
		localizers: localizers,
	}, nil
}

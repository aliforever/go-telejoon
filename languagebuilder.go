package telejoon

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type LanguageConfig struct {
	languages *Languages
	repo      UserLanguageRepository

	forceChooseLanguage           bool
	changeLanguageState           string
	reverseButtonOrderInRowForRTL bool
}

func NewLanguageConfig(languages *Languages, repo UserLanguageRepository) *LanguageConfig {
	return &LanguageConfig{
		languages: languages,
		repo:      repo,
	}
}

// WithChangeLanguageMenu sets the change language menu state.
func (l *LanguageConfig) WithChangeLanguageMenu(state string, forceChooseLanguage bool) *LanguageConfig {
	l.changeLanguageState = state
	l.forceChooseLanguage = forceChooseLanguage

	return l
}

// WithReverseButtonOrderInRowForRTL sets the reverse button order in row for RTL languages.
func (l *LanguageConfig) WithReverseButtonOrderInRowForRTL() *LanguageConfig {
	l.reverseButtonOrderInRowForRTL = true

	return l
}

// GetLanguage By Tag
func (l *LanguageConfig) GetLanguage(tag string) *Language {
	return l.languages.GetByTag(tag)
}

type Language struct {
	tag       string
	rtl       bool
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
func (l *Language) MustGet(id string) string {
	return l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
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
func (l *Languages) GetByTag(tag string) *Language {
	for index := range l.localizers {
		localizer := l.localizers[index]
		if localizer.tag == tag {
			return &localizer
		}
	}

	return nil
}

type LanguagesBuilder struct {
	defaultBundle    *i18n.Bundle
	rtlLanguageTags  []language.Tag
	unmarshalFuncs   map[string]i18n.UnmarshalFunc
	messageFilePaths []string
}

func NewLanguageBuilder(defaultBundle language.Tag, rtlLanguageTags ...language.Tag) *LanguagesBuilder {
	return &LanguagesBuilder{
		defaultBundle:   i18n.NewBundle(defaultBundle),
		rtlLanguageTags: rtlLanguageTags,
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

	for i := range lb.messageFilePaths {
		path := lb.messageFilePaths[i]

		if msgFile, err := lb.defaultBundle.LoadMessageFile(path); err != nil {
			return nil, err
		} else {
			isRtl := false

			for j := range lb.rtlLanguageTags {
				if msgFile.Tag == lb.rtlLanguageTags[j] {
					isRtl = true
					break
				}
			}

			localizers = append(localizers, Language{
				tag:       msgFile.Tag.String(),
				rtl:       isRtl,
				localizer: i18n.NewLocalizer(lb.defaultBundle, msgFile.Tag.String()),
			})
		}
	}

	return &Languages{
		localizers: localizers,
	}, nil
}

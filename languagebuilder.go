package telejoon

import (
	"fmt"
	"io/fs"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// LanguageConfig wires the loaded languages into an engine: how the user's
// language is stored, and optionally the automatic change-language menu.
type LanguageConfig struct {
	languages *Languages
	repo      UserLanguageRepository

	forceChooseLanguage           bool
	changeLanguageState           string
	reverseButtonOrderInRowForRTL bool
}

// NewLanguageConfig creates a language config. A nil repo uses an in-memory
// repository — fine for a single-process bot, but pair a persistent
// UserLanguageRepository with a persistent UserRepository in production.
func NewLanguageConfig(languages *Languages, repo UserLanguageRepository) *LanguageConfig {
	if repo == nil {
		repo = NewDefaultUserLanguageRepository()
	}

	return &LanguageConfig{
		languages: languages,
		repo:      repo,
	}
}

// WithChangeLanguageMenu registers an automatic change-language menu under
// the given state. When forceChooseLanguage is true, users without a chosen
// language are redirected to this menu before anything else.
//
// The menu renders one button per language; each button's label is the
// language's translation of "<StateName>.Button" (falling back to the
// language tag), and the menu text is every language's "<StateName>.Text".
// Choosing a language persists it and lands the user on the default state.
func (l *LanguageConfig) WithChangeLanguageMenu(state State[NoData], forceChooseLanguage bool) *LanguageConfig {
	l.changeLanguageState = state.name
	l.forceChooseLanguage = forceChooseLanguage

	return l
}

// WithReverseButtonOrderInRowForRTL reverses the button order within each
// keyboard row for right-to-left languages.
func (l *LanguageConfig) WithReverseButtonOrderInRowForRTL() *LanguageConfig {
	l.reverseButtonOrderInRowForRTL = true

	return l
}

// GetLanguage returns the language for the given tag, or nil.
func (l *LanguageConfig) GetLanguage(tag string) *Language {
	return l.languages.GetByTag(tag)
}

// Language is a loaded locale: its tag, text direction, and localizer.
type Language struct {
	tag       string
	rtl       bool
	localizer *i18n.Localizer
}

// Tag returns the language's BCP 47 tag (e.g. "en", "fa").
func (l *Language) Tag() string {
	return l.tag
}

// IsRTL reports whether the language is written right-to-left.
func (l *Language) IsRTL() bool {
	return l.rtl
}

// Get returns the localized string for the given message ID.
func (l *Language) Get(id string) (string, error) {
	if l == nil || l.localizer == nil {
		return "", fmt.Errorf("language_not_available: %s", id)
	}

	return l.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: id,
	})
}

// MustGet returns the localized string for the given message ID.
// It panics when the message ID is not translated; a nil Language renders "".
func (l *Language) MustGet(id string) string {
	if l == nil || l.localizer == nil {
		return ""
	}

	return l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
	})
}

// GetWithParams returns the localized string for the given message ID and parameters.
func (l *Language) GetWithParams(id string, params map[string]interface{}) (string, error) {
	if l == nil || l.localizer == nil {
		return "", fmt.Errorf("language_not_available: %s", id)
	}

	return l.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: params,
	})
}

// MustGetWithParams returns the localized string for the given message ID and parameters.
// It panics when the message ID is not translated; a nil Language renders "".
func (l *Language) MustGetWithParams(id string, params map[string]interface{}) string {
	if l == nil || l.localizer == nil {
		return ""
	}

	return l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: params,
	})
}

// Languages is the loaded set of locales. Keys missing from a language fall
// back to the default language; a user with no (or an unknown) chosen
// language renders the default language.
type Languages struct {
	localizers []Language
	byTag      map[string]*Language
	defaultTag string
}

// GetByTag returns the language for the given tag, or nil.
func (l *Languages) GetByTag(tag string) *Language {
	if l == nil {
		return nil
	}

	return l.byTag[tag]
}

// Default returns the default language.
func (l *Languages) Default() *Language {
	if l == nil {
		return nil
	}

	return l.byTag[l.defaultTag]
}

// All returns the loaded languages in load order.
func (l *Languages) All() []Language {
	if l == nil {
		return nil
	}

	return append([]Language(nil), l.localizers...)
}

// LanguagesBuilder loads message files into a Languages set.
//
//	languages, err := telejoon.NewLanguageBuilder(language.English).
//		AddTOML("locale.en.toml", "locale.fa.toml").
//		Build()
//
// Text direction is auto-detected from each language's script (Persian,
// Arabic, Hebrew, … are RTL); use WithRTL to override the detection.
type LanguagesBuilder struct {
	defaultTag  language.Tag
	bundle      *i18n.Bundle
	rtlOverride map[string]bool
	files       []messageFile
}

type messageFile struct {
	fsys fs.FS // nil means the OS filesystem
	path string
}

// NewLanguageBuilder creates a builder whose default (fallback) language is
// defaultTag. Build fails unless a message file for the default language is
// added.
func NewLanguageBuilder(defaultTag language.Tag) *LanguagesBuilder {
	return &LanguagesBuilder{
		defaultTag: defaultTag,
		bundle:     i18n.NewBundle(defaultTag),
	}
}

// AddTOML adds TOML message files loaded from the OS filesystem. The
// language of each file is taken from its content ("en" in locale.en.toml),
// not from the file name.
func (lb *LanguagesBuilder) AddTOML(paths ...string) *LanguagesBuilder {
	lb.bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	for _, path := range paths {
		lb.files = append(lb.files, messageFile{path: path})
	}

	return lb
}

// AddTOMLFS adds TOML message files loaded from fsys (e.g. an embed.FS).
func (lb *LanguagesBuilder) AddTOMLFS(fsys fs.FS, paths ...string) *LanguagesBuilder {
	lb.bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	for _, path := range paths {
		lb.files = append(lb.files, messageFile{fsys: fsys, path: path})
	}

	return lb
}

// WithRTL marks the given languages as right-to-left, in addition to script
// auto-detection. Use it for languages written in a script that is not
// inherently RTL (e.g. an Arabic-language file transliterated into Latin).
func (lb *LanguagesBuilder) WithRTL(tags ...language.Tag) *LanguagesBuilder {
	if lb.rtlOverride == nil {
		lb.rtlOverride = map[string]bool{}
	}

	for _, tag := range tags {
		lb.rtlOverride[tag.String()] = true
	}

	return lb
}

// rtlScripts are ISO 15924 script codes written right-to-left.
var rtlScripts = map[string]bool{
	"Arab": true, // Arabic (incl. Persian, Urdu)
	"Aran": true, // Arabic Nastaliq
	"Hebr": true, // Hebrew
	"Syrc": true, // Syriac
	"Thaa": true, // Thaana (Dhivehi)
	"Nkoo": true, // N'Ko
	"Samr": true, // Samaritan
	"Mand": true, // Mandaic
	"Adlm": true, // Adlam
	"Rohg": true, // Hanifi Rohingya
	"Mero": true, // Meroitic
}

// isRTLTag reports whether the tag's script is written right-to-left.
func isRTLTag(tag language.Tag) bool {
	script, confidence := tag.Script()

	return confidence >= language.High && rtlScripts[script.String()]
}

// Build loads the message files and returns the language set.
func (lb *LanguagesBuilder) Build() (*Languages, error) {
	languages := &Languages{
		byTag:      map[string]*Language{},
		defaultTag: lb.defaultTag.String(),
	}

	seen := map[string]string{}

	for _, file := range lb.files {
		var (
			msgFile *i18n.MessageFile
			err     error
		)

		if file.fsys != nil {
			msgFile, err = lb.bundle.LoadMessageFileFS(file.fsys, file.path)
		} else {
			msgFile, err = lb.bundle.LoadMessageFile(file.path)
		}

		if err != nil {
			return nil, err
		}

		tag := msgFile.Tag.String()

		if first, duplicate := seen[tag]; duplicate {
			return nil, fmt.Errorf("duplicate_language_file: %s (%s and %s)", tag, first, file.path)
		}

		seen[tag] = file.path

		rtl := isRTLTag(msgFile.Tag) || lb.rtlOverride[tag]

		// Fall back to the bundle's default language for keys missing in this
		// language, instead of erroring (and panicking via MustGet).
		localizerTags := []string{tag}
		if languages.defaultTag != tag {
			localizerTags = append(localizerTags, languages.defaultTag)
		}

		languages.localizers = append(languages.localizers, Language{
			tag:       tag,
			rtl:       rtl,
			localizer: i18n.NewLocalizer(lb.bundle, localizerTags...),
		})
	}

	if len(languages.localizers) == 0 {
		return nil, fmt.Errorf("no_language_files_added")
	}

	for i := range languages.localizers {
		lang := languages.localizers[i]
		languages.byTag[lang.tag] = &lang
	}

	if languages.byTag[languages.defaultTag] == nil {
		return nil, fmt.Errorf("no_message_file_for_default_language: %s", languages.defaultTag)
	}

	return languages, nil
}

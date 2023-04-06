package telejoon

type LanguageI interface {
	Flag() string
	Code() string
	Name() string
	SelectLanguage() string
	Welcome() string
}

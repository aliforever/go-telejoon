package telejoon

type TextBuilder interface {
	String(update *StateUpdate) string
}

type DeferredTextBuilder func(update *StateUpdate) string

func (t DeferredTextBuilder) String(update *StateUpdate) string {
	return t(update)
}

type StaticTextBuilder string

func (t StaticTextBuilder) String(_ *StateUpdate) string {
	return string(t)
}

type LanguageKeyTextBuilder string

func (t LanguageKeyTextBuilder) String(update *StateUpdate) string {
	return update.Language().MustGet(string(t))
}

// NewLanguageKeyText returns a new LanguageKeyTextBuilder
func NewLanguageKeyText(key string) LanguageKeyTextBuilder {
	return LanguageKeyTextBuilder(key)
}

// NewStaticText returns a new StaticTextBuilder
func NewStaticText(text string) StaticTextBuilder {
	return StaticTextBuilder(text)
}

// NewDeferredText returns a new DeferredTextBuilder
func NewDeferredText(text func(update *StateUpdate) string) DeferredTextBuilder {
	return text
}

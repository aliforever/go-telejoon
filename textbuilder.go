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

type LanguageKeyWithParamsTextBuilder struct {
	key    string
	params map[string]interface{}
}

func (t LanguageKeyWithParamsTextBuilder) String(update *StateUpdate) string {
	return update.Language().MustGetWithParams(t.key, t.params)
}

type UpdateKeyTextBuilder string

func (t UpdateKeyTextBuilder) String(update *StateUpdate) string {
	data := update.Get(string(t))

	if data == nil {
		return ""
	}

	if val, ok := data.(string); ok {
		return val
	}

	return ""
}

// NewUpdateKeyText returns a new UpdateKeyTextBuilder
func NewUpdateKeyText(key string) UpdateKeyTextBuilder {
	return UpdateKeyTextBuilder(key)
}

// NewLanguageKeyWithParamsText returns a new LanguageKeyWithParamsTextBuilder
func NewLanguageKeyWithParamsText(key string, params map[string]interface{}) LanguageKeyWithParamsTextBuilder {
	return LanguageKeyWithParamsTextBuilder{
		key:    key,
		params: params,
	}
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

package telejoon

type ButtonOptions struct {
	translateName bool
	translateText bool
}

// NewButtonOptions creates a new ButtonOptions.
func NewButtonOptions() *ButtonOptions {
	return &ButtonOptions{
		translateName: true,
		translateText: true,
	}
}

// TranslateName sets whether to translate the title of the action.
func (b *ButtonOptions) TranslateName() *ButtonOptions {
	b.translateName = true
	return b
}

// TranslateText sets whether to translate the textHandler of the action.
func (b *ButtonOptions) TranslateText() *ButtonOptions {
	b.translateText = true
	return b
}

// DoNotTranslateName sets whether to translate the title of the action.
func (b *ButtonOptions) DoNotTranslateName() *ButtonOptions {
	b.translateName = false
	return b
}

// DoNotTranslateText sets whether to translate the textHandler of the action.
func (b *ButtonOptions) DoNotTranslateText() *ButtonOptions {
	b.translateText = false
	return b
}

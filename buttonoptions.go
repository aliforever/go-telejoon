package telejoon

type ButtonOptions struct {
	translateTitle bool
	translateText  bool
}

// NewButtonOptions creates a new ButtonOptions.
func NewButtonOptions() *ButtonOptions {
	return &ButtonOptions{
		translateTitle: true,
		translateText:  true,
	}
}

// TranslateTitle sets whether to translate the title of the action.
func (b *ButtonOptions) TranslateTitle() *ButtonOptions {
	b.translateTitle = true
	return b
}

// TranslateText sets whether to translate the text of the action.
func (b *ButtonOptions) TranslateText() *ButtonOptions {
	b.translateText = true
	return b
}

// DoNotTranslateTitle sets whether to translate the title of the action.
func (b *ButtonOptions) DoNotTranslateTitle() *ButtonOptions {
	b.translateTitle = false
	return b
}

// DoNotTranslateText sets whether to translate the text of the action.
func (b *ButtonOptions) DoNotTranslateText() *ButtonOptions {
	b.translateText = false
	return b
}

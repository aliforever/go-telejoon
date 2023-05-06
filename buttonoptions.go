package telejoon

type ButtonOptions struct {
	translateName bool
	translateText bool
	breakBefore   bool
	breakAfter    bool
}

// NewButtonOptions creates a new ButtonOptions.
func NewButtonOptions() *ButtonOptions {
	return &ButtonOptions{
		translateName: false,
		translateText: false,
	}
}

// BreakBefore sets whether to break before the button when building keyboard.
func (b *ButtonOptions) BreakBefore() *ButtonOptions {
	b.breakBefore = true
	return b
}

// BreakAfter sets whether to break after the button when building keyboard.
func (b *ButtonOptions) BreakAfter() *ButtonOptions {
	b.breakAfter = true
	return b
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

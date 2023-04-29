package telejoon

type ButtonOptions struct {
	translateName bool
	translateText bool
	breakBefore   bool
	breakAfter    bool
	// shouldEdit    determines if the message should be edited or not, used when switching to a new state by sending .
	//	a new message or editing the current message.
	shouldEdit bool
	// alert determines if the alert should be shown when the button is clicked. Only used in callbacks.
	alert           string
	showAlertDialog bool
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

// ShouldEdit sets whether to edit the message when the button is clicked.
func (b *ButtonOptions) ShouldEdit() *ButtonOptions {
	b.shouldEdit = true
	return b
}

// ShouldNotEdit sets whether to edit the message when the button is clicked.
func (b *ButtonOptions) ShouldNotEdit() *ButtonOptions {
	b.shouldEdit = false
	return b
}

// Alert sets the alert for the button.
func (b *ButtonOptions) Alert(alert string) *ButtonOptions {
	b.alert = alert
	return b
}

// DisableAlert disables the alert for the button.
func (b *ButtonOptions) DisableAlert() *ButtonOptions {
	b.alert = ""
	return b
}

// ShowAlertDialog sets whether to show the alert dialog when the button is clicked.
func (b *ButtonOptions) ShowAlertDialog() *ButtonOptions {
	b.showAlertDialog = true
	return b
}

// DoNotShowAlertDialog sets whether to show the alert dialog when the button is clicked.
func (b *ButtonOptions) DoNotShowAlertDialog() *ButtonOptions {
	b.showAlertDialog = false
	return b
}

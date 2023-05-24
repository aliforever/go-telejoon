package telejoon

type ButtonOptions struct {
	breakBefore bool
	breakAfter  bool
}

// NewButtonOptions creates a new ButtonOptions.
func NewButtonOptions(breakBefore, BreakAfter bool) *ButtonOptions {
	return &ButtonOptions{
		breakBefore: breakBefore,
		breakAfter:  BreakAfter,
	}
}

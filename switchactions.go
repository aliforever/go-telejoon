package telejoon

type SwitchAction interface {
	target() string
}

type SwitchActionInlineMenu struct {
	targetInlineMenu string
	edit             bool
}

func (s *SwitchActionInlineMenu) target() string {
	return s.targetInlineMenu
}

type SwitchActionState struct {
	targetState string
}

func (s *SwitchActionState) target() string {
	return s.targetState
}

// NewSwitchActionInlineMenu creates a new SwitchActionInlineMenu
func NewSwitchActionInlineMenu(targetInlineMenu string, edit bool) *SwitchActionInlineMenu {
	return &SwitchActionInlineMenu{targetInlineMenu: targetInlineMenu, edit: edit}
}

// NewSwitchActionState creates a new SwitchActionState
func NewSwitchActionState(targetState string) *SwitchActionState {
	return &SwitchActionState{targetState: targetState}
}

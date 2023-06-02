package telejoon

type SwitchAction interface {
	target() interface{}
}

type SwitchActionInlineMenu struct {
	targetInlineMenu *InlineMenu
	edit             bool
}

func (s *SwitchActionInlineMenu) target() *InlineMenu {
	return s.targetInlineMenu
}

type SwitchActionState struct {
	targetState *StaticMenu
}

func (s *SwitchActionState) target() *StaticMenu {
	return s.targetState
}

// NewSwitchActionInlineMenu creates a new SwitchActionInlineMenu
func NewSwitchActionInlineMenu(inlineMenu *InlineMenu, edit bool) *SwitchActionInlineMenu {
	return &SwitchActionInlineMenu{targetInlineMenu: inlineMenu, edit: edit}
}

// NewSwitchActionState creates a new SwitchActionState
func NewSwitchActionState(targetState *StaticMenu) *SwitchActionState {
	return &SwitchActionState{targetState: targetState}
}

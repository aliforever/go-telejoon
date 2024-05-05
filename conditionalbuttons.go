package telejoon

type conditionalButtons struct {
	cond func(update *StateUpdate) bool

	buttons []Action

	formation []int
}

func (b *ActionBuilder) AddConditionalButtons(
	cond func(update *StateUpdate) bool,
	buttons ...baseButton,
) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(buttons) == 0 {
		return b
	}

	btns := make([]Action, len(buttons))

	for i := range buttons {
		btns[i] = buttons[i]
	}

	b.conditionalButtons = append(b.conditionalButtons, conditionalButtons{
		cond:      cond,
		buttons:   btns,
		formation: b.buttonFormation,
	})

	return b
}

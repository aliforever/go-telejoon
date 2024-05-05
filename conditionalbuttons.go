package telejoon

type conditionalButtons struct {
	cond             func(update *StateUpdate) bool
	cachedCondResult *bool

	buttons []Action

	formation []int
}

func (b *ActionBuilder) AddConditionalButtons(
	cond func(update *StateUpdate) bool,
	buttonFormation []int,
	buttons ...Action,
) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(buttons) == 0 {
		return b
	}

	b.conditionalButtons = append(b.conditionalButtons, conditionalButtons{
		cond: cond,

		buttons:   buttons,
		formation: buttonFormation,
	})

	return b
}

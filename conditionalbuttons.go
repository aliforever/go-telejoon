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

	cb := conditionalButtons{
		buttons:   buttons,
		formation: buttonFormation,
	}

	cb.cond = func(update *StateUpdate) bool {
		if cb.cachedCondResult != nil {
			return *cb.cachedCondResult
		}

		condResult := cond(update)

		cb.cachedCondResult = &condResult

		return condResult
	}

	b.conditionalButtons = append(b.conditionalButtons, cb)

	return b
}

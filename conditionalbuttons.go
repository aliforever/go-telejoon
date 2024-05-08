package telejoon

type conditionalButtons struct {
	definedCondition   *string
	vsDefinedCondition *string
	cond               func(update *StateUpdate) bool

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

func (b *ActionBuilder) AddDefinedConditionalButtons(
	definedCondition string,
	buttonFormation []int,
	buttons ...Action,
) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(buttons) == 0 {
		return b
	}

	b.conditionalButtons = append(b.conditionalButtons, conditionalButtons{
		definedCondition: &definedCondition,

		buttons:   buttons,
		formation: buttonFormation,
	})

	return b
}

func (b *ActionBuilder) AddVsDefinedConditionalButtons(
	vsDefinedCondition string,
	buttonFormation []int,
	buttons ...Action,
) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(buttons) == 0 {
		return b
	}

	b.conditionalButtons = append(b.conditionalButtons, conditionalButtons{
		vsDefinedCondition: &vsDefinedCondition,

		buttons:   buttons,
		formation: buttonFormation,
	})

	return b
}

func (b *ActionBuilder) getConditionalButtonByName(
	update *StateUpdate,
	name string,
) Action {

	if len(b.conditionalButtons) == 0 {
		return nil
	}

	for _, acb := range b.conditionalButtons {
		condition1 := len(b.definedConditionResults) == 0 ||
			(acb.definedCondition != nil && b.definedConditionResults[*acb.definedCondition])

		condition2 := acb.cond != nil && acb.cond(update)

		if condition1 && condition2 {
			for _, action := range acb.buttons {
				if action.Name(update) == name {
					if opts, ok := action.(baseButtonOptions); ok {
						if opts.CanBeShown(update, b.definedConditionResults) {
							return action
						}
					} else {
						return action
					}
				}
			}
		}
	}

	return nil
}

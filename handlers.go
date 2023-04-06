package telejoon

type Handlers[User UserI, Lang LanguageI] struct {
	stateHandlers    *StateHandlers[User, Lang]
	callbackHandlers *CallbackHandlers[User, Lang]
}

func NewHandlers[User UserI, Lang LanguageI]() *Handlers[User, Lang] {
	return &Handlers[User, Lang]{}
}

func (h *Handlers[User, Lang]) SetStateHandlers(s *StateHandlers[User, Lang]) *Handlers[User, Lang] {
	h.stateHandlers = s

	return h
}

func (h *Handlers[User, Lang]) SetCallbackHandlers(c *CallbackHandlers[User, Lang]) *Handlers[User, Lang] {
	h.callbackHandlers = c

	return h
}

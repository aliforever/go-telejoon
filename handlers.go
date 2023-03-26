package telejoon

type Handlers[User any] struct {
	stateHandlers    *StateHandlers[User]
	callbackHandlers *CallbackHandlers[User]
}

func NewHandlers[User any]() *Handlers[User] {
	return &Handlers[User]{}
}

func (h *Handlers[User]) SetStateHandlers(s *StateHandlers[User]) *Handlers[User] {
	h.stateHandlers = s

	return h
}

func (h *Handlers[User]) SetCallbackHandlers(c *CallbackHandlers[User]) *Handlers[User] {
	h.callbackHandlers = c

	return h
}

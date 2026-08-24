package telejoon

import (
	"fmt"

	"github.com/aliforever/go-telegram-bot-api/structs"
)

// Part is the set of message payloads a menu can react to with On.
type Part interface {
	[]structs.PhotoSize | *structs.Video | *structs.Document | *structs.Voice |
		*structs.Audio | *structs.Sticker | *structs.Location | *structs.Contact |
		*structs.VideoNote | *structs.Venue | *structs.Poll | *structs.Dice
}

type partKind int

const (
	partPhoto partKind = iota
	partVideo
	partDocument
	partVoice
	partAudio
	partSticker
	partLocation
	partContact
	partVideoNote
	partVenue
	partPoll
	partDice
)

// partKindOf maps a Part type parameter to its runtime kind. It is the single
// exhaustive switch mirroring the Part union above.
func partKindOf[P Part]() partKind {
	var zero P

	switch any(zero).(type) {
	case []structs.PhotoSize:
		return partPhoto
	case *structs.Video:
		return partVideo
	case *structs.Document:
		return partDocument
	case *structs.Voice:
		return partVoice
	case *structs.Audio:
		return partAudio
	case *structs.Sticker:
		return partSticker
	case *structs.Location:
		return partLocation
	case *structs.Contact:
		return partContact
	case *structs.VideoNote:
		return partVideoNote
	case *structs.Venue:
		return partVenue
	case *structs.Poll:
		return partPoll
	case *structs.Dice:
		return partDice
	}

	panic(fmt.Sprintf("unreachable: unsupported Part type %T", zero))
}

// extractPart pulls the payload of type P out of a message.
func extractPart[P Part](msg *structs.Message) P {
	var part P

	switch out := any(&part).(type) {
	case *[]structs.PhotoSize:
		*out = msg.Photo
	case **structs.Video:
		*out = msg.Video
	case **structs.Document:
		*out = msg.Document
	case **structs.Voice:
		*out = msg.Voice
	case **structs.Audio:
		*out = msg.Audio
	case **structs.Sticker:
		*out = msg.Sticker
	case **structs.Location:
		*out = msg.Location
	case **structs.Contact:
		*out = msg.Contact
	case **structs.VideoNote:
		*out = msg.VideoNote
	case **structs.Venue:
		*out = msg.Venue
	case **structs.Poll:
		*out = msg.Poll
	case **structs.Dice:
		*out = msg.Dice
	}

	return part
}

// detectPart returns the part kind carried by a message, or false.
func detectPart(msg *structs.Message) (partKind, bool) {
	switch {
	case msg.Video != nil:
		return partVideo, true
	case msg.Photo != nil:
		return partPhoto, true
	case msg.Document != nil:
		return partDocument, true
	case msg.Voice != nil:
		return partVoice, true
	case msg.Audio != nil:
		return partAudio, true
	case msg.Sticker != nil:
		return partSticker, true
	case msg.Location != nil:
		return partLocation, true
	case msg.Contact != nil:
		return partContact, true
	case msg.VideoNote != nil:
		return partVideoNote, true
	case msg.Venue != nil:
		return partVenue, true
	case msg.Poll != nil:
		return partPoll, true
	case msg.Dice != nil:
		return partDice, true
	}

	return 0, false
}

// MenuBuilder builds the state menu bound to a State[D]. Handlers receive the
// state's resolved payload as data *D. Compile it into the engine with
// engine.Add.
type MenuBuilder[D any] struct {
	state State[D]
	text  Text

	buttons   []*Button
	buttonsFn func(ctx *Ctx, data *D) []*Button

	onTextHandler  func(ctx *Ctx, data *D, text string) Action
	defaultHandler func(ctx *Ctx, data *D) Action
	partHandlers   map[partKind]func(ctx *Ctx, data *D, msg *structs.Message) Action

	middlewares []Handler

	formation []int
	maxPerRow int
}

// Menu creates a builder for the state menu bound to state.
func Menu[D any](state State[D], text Text) *MenuBuilder[D] {
	return &MenuBuilder[D]{
		state:        state,
		text:         text,
		partHandlers: map[partKind]func(*Ctx, *D, *structs.Message) Action{},
	}
}

// Buttons sets the static reply-keyboard buttons of the menu.
func (b *MenuBuilder[D]) Buttons(buttons ...*Button) *MenuBuilder[D] {
	b.buttons = append(b.buttons, buttons...)

	return b
}

// ButtonsFunc sets a per-request reply-keyboard button factory.
func (b *MenuBuilder[D]) ButtonsFunc(fn func(ctx *Ctx, data *D) []*Button) *MenuBuilder[D] {
	b.buttonsFn = fn

	return b
}

// On registers a handler for one message part type. P is inferred from the
// closure's parameter — the caller never writes a type argument:
//
//	menu.On(func(ctx *telejoon.Ctx, data *telejoon.NoData, photo []structs.PhotoSize) telejoon.Action {
//		return ctx.ReplyText("nice photo!")
//	})
func (b *MenuBuilder[D]) On[P Part](fn func(ctx *Ctx, data *D, part P) Action) *MenuBuilder[D] {
	b.partHandlers[partKindOf[P]()] = func(ctx *Ctx, data *D, msg *structs.Message) Action {
		return fn(ctx, data, extractPart[P](msg))
	}

	return b
}

// OnText registers the handler for text messages that did not match any
// visible button label.
func (b *MenuBuilder[D]) OnText(fn func(ctx *Ctx, data *D, text string) Action) *MenuBuilder[D] {
	b.onTextHandler = fn

	return b
}

// Default registers the fallback handler for messages with no matching
// button, text handler, or part handler.
func (b *MenuBuilder[D]) Default(fn func(ctx *Ctx, data *D) Action) *MenuBuilder[D] {
	b.defaultHandler = fn

	return b
}

// Use adds a middleware that runs before any handler of this menu.
// Return Next to continue, anything else to stop.
func (b *MenuBuilder[D]) Use(middleware Handler) *MenuBuilder[D] {
	b.middlewares = append(b.middlewares, middleware)

	return b
}

// Formation sets the reply-keyboard row sizes (see chunkIntoRows).
func (b *MenuBuilder[D]) Formation(formation ...int) *MenuBuilder[D] {
	b.formation = formation

	return b
}

// MaxPerRow caps the number of reply-keyboard buttons per row.
func (b *MenuBuilder[D]) MaxPerRow(max int) *MenuBuilder[D] {
	b.maxPerRow = max

	return b
}

// register compiles the builder into the engine. It is a non-generic method,
// so MenuBuilder[D] satisfies the Registrable interface for any D.
func (b *MenuBuilder[D]) register(e *Engine) error {
	if !validRouteName(b.state.name) {
		return fmt.Errorf("invalid_state_name: %q", b.state.name)
	}

	if _, exists := e.menus[b.state.name]; exists {
		return fmt.Errorf("duplicate_state_menu: %s", b.state.name)
	}

	e.menus[b.state.name] = b.compile()

	return nil
}

// compile lowers the typed builder into the erased runtime representation.
// All type assertions on D happen behind this boundary.
func (b *MenuBuilder[D]) compile() *menuRuntime {
	runtime := &menuRuntime{
		state:         b.state.name,
		text:          b.text,
		middlewares:   b.middlewares,
		formation:     b.formation,
		maxPerRow:     b.maxPerRow,
		parts:         map[partKind]func(*Ctx) Action{},
		staticButtons: b.buttons,
	}

	runtime.loadData = func(ctx *Ctx) any {
		// A GoToWith transition during this request carries the payload
		// directly; the types match by construction. The payload is copied so
		// handlers never share a *D across users or requests.
		pending := ctx.pendingData
		ctx.pendingData = nil

		if pending != nil {
			if data, ok := pending.(*D); ok {
				copied := *data

				return &copied
			}

			ctx.engine.onErr(ctx.Context(), ctx.client, ctx.Update,
				fmt.Errorf("state_data_type_mismatch: %s", b.state.name))
		}

		data := new(D)

		if repo := ctx.engine.getStateDataRepository(); repo != nil {
			raw, err := repo.GetUserStateData(ctx.UserID(), b.state.name)

			switch {
			case err != nil:
				ctx.engine.onErr(ctx.Context(), ctx.client, ctx.Update,
					fmt.Errorf("state_data_load: %s: %w", b.state.name, err))
			case len(raw) > 0:
				if err := unmarshalStateData(raw, data); err != nil {
					ctx.engine.onErr(ctx.Context(), ctx.client, ctx.Update,
						fmt.Errorf("state_data_decode: %s: %w", b.state.name, err))
				}
			}
		}

		return data
	}

	runtime.keyboard = func(ctx *Ctx) ([]*Button, error) {
		if b.buttonsFn != nil {
			data, ok := ctx.stateData.(*D)
			if !ok {
				return nil, fmt.Errorf("state_data_type_mismatch: %s", b.state.name)
			}

			return b.buttonsFn(ctx, data), nil
		}

		return b.buttons, nil
	}

	if b.onTextHandler != nil {
		fn := b.onTextHandler
		runtime.onText = func(ctx *Ctx) Action {
			return fn(ctx, ctx.stateData.(*D), ctx.Text())
		}
	}

	if b.defaultHandler != nil {
		fn := b.defaultHandler
		runtime.onDefault = func(ctx *Ctx) Action {
			return fn(ctx, ctx.stateData.(*D))
		}
	}

	for kind, handler := range b.partHandlers {
		fn := handler
		runtime.parts[kind] = func(ctx *Ctx) Action {
			return fn(ctx, ctx.stateData.(*D), ctx.Update.Message)
		}
	}

	return runtime
}

package telejoon

import (
	"fmt"
	"net/url"

	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
)

// Registrable is anything Engine.Add can compile into the engine:
// *MenuBuilder[D] (for any D) or *InlineMenuBuilder. The interface is sealed;
// generic builders satisfy it with their non-generic register method.
type Registrable interface {
	register(e *Engine) error
}

// menuRuntime is the erased, immutable runtime form of a state menu.
// The generic parameters of the originating builder are captured in
// closures; the dispatch loop stays monomorphic.
type menuRuntime struct {
	state string

	text Text

	// keyboard resolves the visible buttons for the request.
	keyboard func(ctx *Ctx) ([]*Button, error)

	// loadData resolves the state payload for the request, as a *D.
	loadData func(ctx *Ctx) any

	onText    func(ctx *Ctx) Action
	onDefault func(ctx *Ctx) Action
	parts     map[partKind]func(ctx *Ctx) Action

	middlewares []Handler

	formation []int
	maxPerRow int

	// staticButtons is kept for Validate's cross-reference checks.
	staticButtons []*Button
}

// inlineMenuRuntime is the erased runtime form of an inline menu.
type inlineMenuRuntime struct {
	name string

	text Text

	buttons   []*InlineButton
	buttonsFn func(ctx *Ctx) []*InlineButton

	routes map[string]erasedRouteHandler

	middlewares []Handler

	formation []int
	maxPerRow int
}

// renderReplyKeyboard renders the reply keyboard once per request and returns
// the markup together with the label→button dispatch map. Dispatching through
// the same labels that were rendered keeps rendering and dispatch consistent;
// duplicate visible labels in the user's language are an explicit error.
//
// The dispatch map additionally contains the labels rendered in every other
// configured language, so a user who switched language mid-state still
// matches the previously-sent keyboard.
func (e *Engine) renderReplyKeyboard(
	ctx *Ctx,
	runtime *menuRuntime,
) (*structs.ReplyKeyboardMarkup, map[string]*Button, error) {

	buttons, err := runtime.keyboard(ctx)
	if err != nil {
		return nil, nil, err
	}

	var visible []*Button

	for _, button := range buttons {
		if button != nil && button.canBeShown(ctx) {
			visible = append(visible, button)
		}
	}

	if len(visible) == 0 {
		return nil, nil, nil
	}

	byLabel := map[string]*Button{}

	for _, button := range visible {
		label := button.label(ctx)
		if _, duplicate := byLabel[label]; duplicate {
			return nil, nil, fmt.Errorf("duplicate_button_label: %q in state %s", label, runtime.state)
		}

		byLabel[label] = button
	}

	// Labels in the other configured languages, for cross-language dispatch.
	// Conflicts across languages are resolved first-come; two languages may
	// legitimately translate distinct buttons identically.
	if e.languageConfig != nil && len(e.languageConfig.languages.localizers) > 1 {
		for i := range e.languageConfig.languages.localizers {
			lang := e.languageConfig.languages.localizers[i]

			if ctx.language != nil && lang.tag == ctx.language.tag {
				continue
			}

			clone := *ctx
			clone.language = &lang

			for _, button := range visible {
				label := button.label(&clone)
				if _, exists := byLabel[label]; !exists {
					byLabel[label] = button
				}
			}
		}
	}

	var labels []string

	for _, button := range visible {
		if button.breakBefore {
			labels = append(labels, "")
		}

		labels = append(labels, button.label(ctx))

		if button.breakAfter {
			labels = append(labels, "")
		}
	}

	reverse := ctx.language != nil && ctx.language.rtl &&
		e.languageConfig != nil && e.languageConfig.reverseButtonOrderInRowForRTL

	markup := tools.Keyboards{}.NewReplyKeyboardFromSlicesOfStrings(
		chunkIntoRows(labels, isEmptyString, runtime.maxPerRow, runtime.formation, reverse),
	)

	return markup, byLabel, nil
}

// buildInlineMarkup renders the inline keyboard for the request, encoding
// callback_data for each button kind and enforcing Telegram's 64-byte limit.
func (e *Engine) buildInlineMarkup(
	ctx *Ctx,
	runtime *inlineMenuRuntime,
) (*structs.InlineKeyboardMarkup, error) {

	buttons := runtime.buttons
	if runtime.buttonsFn != nil {
		buttons = runtime.buttonsFn(ctx)
	}

	callbackData := func(data string) (string, error) {
		if len(data) > maxCallbackDataBytes {
			return "", fmt.Errorf(
				"callback_data_too_long: inline menu %s button exceeds %d bytes",
				runtime.name, maxCallbackDataBytes)
		}

		return data, nil
	}

	var rows []map[string]string

	// A button that fails to encode (e.g. oversized callback_data) is
	// skipped and reported, not fatal: the rest of the keyboard still renders.
	var firstErr error

	for _, button := range buttons {
		if button == nil || !button.canBeShown(ctx) {
			continue
		}

		if button.encodeErr != nil {
			if firstErr == nil {
				firstErr = button.encodeErr
			}

			continue
		}

		row := map[string]string{"text": button.label(ctx)}

		data := ""

		switch button.kind {
		case inlineKindURL:
			row["url"] = button.url(ctx)
		case inlineKindCallback:
			data = button.callbackData
		case inlineKindAlert:
			showAlert := "0"
			if button.showAlert {
				showAlert = "1"
			}

			data = runtime.name + ":@a:" + showAlert + ":" + url.QueryEscape(button.alertText)
		case inlineKindMenu:
			if !validRouteName(button.menu) {
				if firstErr == nil {
					firstErr = fmt.Errorf("invalid_inline_menu_target: %q", button.menu)
				}

				continue
			}

			op := "@m"
			if button.edit {
				op = "@e"
			}

			data = runtime.name + ":" + op + ":" + button.menu
		case inlineKindState:
			if !validRouteName(button.state) {
				if firstErr == nil {
					firstErr = fmt.Errorf("invalid_state_target: %q", button.state)
				}

				continue
			}

			data = runtime.name + ":@s:" + button.state
		}

		if button.kind != inlineKindURL {
			checked, err := callbackData(data)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}

				continue
			}

			row["callback_data"] = checked
		}

		if button.breakBefore {
			rows = append(rows, nil)
		}

		rows = append(rows, row)

		if button.breakAfter {
			rows = append(rows, nil)
		}
	}

	if len(rows) == 0 {
		return nil, firstErr
	}

	reverse := ctx.language != nil && ctx.language.rtl &&
		e.languageConfig != nil && e.languageConfig.reverseButtonOrderInRowForRTL

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMaps(
		chunkIntoRows(rows, isNilRow, runtime.maxPerRow, runtime.formation, reverse),
	), firstErr
}

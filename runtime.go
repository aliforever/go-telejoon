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

	// loadData resolves the state payload for the request, as a *D. A
	// non-nil error aborts processing of the update: handlers must never
	// receive a silent zero payload after a repository or decode failure.
	loadData func(ctx *Ctx) (any, error)

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
// the same labels that were rendered keeps rendering and dispatch consistent.
//
// A duplicate visible label in the user's language is ambiguous to dispatch:
// the button is skipped and reported, but the rest of the keyboard still
// renders and dispatches — one bad button must not brick the whole state.
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

	// Labels are rendered once per button and reused for both the dispatch
	// map and the markup.
	var kept []*Button
	var labels []string

	for _, button := range visible {
		label := button.label(ctx)
		if _, duplicate := byLabel[label]; duplicate {
			e.onErr(ctx.Context(), ctx.client, ctx.Update,
				fmt.Errorf("duplicate_button_label: %q in state %s", label, runtime.state))

			continue
		}

		byLabel[label] = button
		kept = append(kept, button)
		labels = append(labels, label)
	}

	if len(kept) == 0 {
		return nil, byLabel, nil
	}

	languageConfig := e.getLanguageConfig()

	// Labels in the other configured languages, for cross-language dispatch.
	// Conflicts across languages are resolved first-come; two languages may
	// legitimately translate distinct buttons identically.
	if languageConfig != nil && len(languageConfig.languages.localizers) > 1 {
		for i := range languageConfig.languages.localizers {
			lang := languageConfig.languages.localizers[i]

			if ctx.language != nil && lang.tag == ctx.language.tag {
				continue
			}

			clone := *ctx
			clone.language = &lang

			for _, button := range kept {
				label := button.label(&clone)
				if _, exists := byLabel[label]; !exists {
					byLabel[label] = button
				}
			}
		}
	}

	var rowLabels []string

	for i, button := range kept {
		if button.breakBefore {
			rowLabels = append(rowLabels, "")
		}

		rowLabels = append(rowLabels, labels[i])

		if button.breakAfter {
			rowLabels = append(rowLabels, "")
		}
	}

	reverse := ctx.language != nil && ctx.language.rtl &&
		languageConfig != nil && languageConfig.reverseButtonOrderInRowForRTL

	markup := tools.Keyboards{}.NewReplyKeyboardFromSlicesOfStrings(
		chunkIntoRows(rowLabels, isEmptyString, runtime.maxPerRow, runtime.formation, reverse),
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

	reverse := false
	if cfg := e.getLanguageConfig(); cfg != nil && ctx.language != nil {
		reverse = ctx.language.rtl && cfg.reverseButtonOrderInRowForRTL
	}

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMaps(
		chunkIntoRows(rows, isNilRow, runtime.maxPerRow, runtime.formation, reverse),
	), firstErr
}

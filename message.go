package telejoon

import (
	"encoding/json"
	"sync"
)

// Msg is a typed handle to a localized message without template params.
// Declare messages as package variables, next to your states:
//
//	var WelcomeText = telejoon.NewMsg("Welcome.Main")
//
//	telejoon.Menu(Welcome, WelcomeText.T())
//
// Declaring handles instead of scattering raw key strings gives you a single
// place to rename a key, and lets engine.Validate check every declared
// message against the loaded locales at startup — a typo becomes a startup
// error instead of a raw key leaking into a chat.
type Msg struct {
	key string
}

// NewMsg declares a localized message handle for the given i18n message ID.
func NewMsg(key string) Msg {
	registerMsgKey(key)

	return Msg{key: key}
}

// Key returns the underlying i18n message ID.
func (m Msg) Key() string {
	return m.key
}

// T returns the Text rendering this message in the request's language.
func (m Msg) T() Text {
	return L(m.key)
}

// MsgP is a typed handle to a localized message with template params. The
// type parameter is the params struct; its JSON field names become the
// template variable names:
//
//	type GreetParams struct {
//		Name string `json:"Name"` // {{.Name}} in the locale file
//	}
//
//	var Greet = telejoon.NewMsgP[GreetParams]("Welcome.Greet")
//
//	telejoon.Menu(Welcome, Greet.T(GreetParams{Name: ctx.FirstName()}))
//
// The params type is part of the handle, so a template and its data can
// never drift apart silently the way map[string]interface{} can.
type MsgP[P any] struct {
	key string
}

// NewMsgP declares a localized message handle with typed template params.
func NewMsgP[P any](key string) MsgP[P] {
	return MsgP[P]{key: key}
}

// Key returns the underlying i18n message ID.
func (m MsgP[P]) Key() string {
	return m.key
}

// T returns the Text rendering this message with the given params.
func (m MsgP[P]) T(params P) Text {
	key := m.key

	return func(ctx *Ctx) string {
		if lang := ctx.resolveLanguage(); lang != nil {
			if text, _ := lang.GetWithParams(key, msgParamsToMap(params)); text != "" {
				return text
			}
		}

		return key
	}
}

// msgParamsToMap converts a params struct to the map go-i18n templates
// consume, via a JSON round-trip: the struct's JSON field names are the
// template variable names.
func msgParamsToMap[P any](params P) map[string]interface{} {
	raw, err := json.Marshal(params)
	if err != nil {
		return map[string]interface{}{}
	}

	out := map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]interface{}{}
	}

	return out
}

// === Message key registry (for startup validation) ===

var msgRegistry = struct {
	sync.Mutex
	keys map[string]struct{}
}{keys: map[string]struct{}{}}

func registerMsgKey(key string) {
	msgRegistry.Lock()
	defer msgRegistry.Unlock()

	msgRegistry.keys[key] = struct{}{}
}

// validateMsgs reports declared message keys (NewMsg) that do not localize
// with the default language. MsgP handles are excluded: their templates
// require params, so they cannot be rendered for a check.
func (l *Languages) validateMsgs() []string {
	msgRegistry.Lock()
	keys := make([]string, 0, len(msgRegistry.keys))
	for key := range msgRegistry.keys {
		keys = append(keys, key)
	}
	msgRegistry.Unlock()

	def := l.Default()
	if def == nil {
		return nil
	}

	var missing []string

	for _, key := range keys {
		if text, err := def.Get(key); err != nil || text == "" {
			missing = append(missing, key)
		}
	}

	return missing
}

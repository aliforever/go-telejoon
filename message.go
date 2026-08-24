package telejoon

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
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
	registerMsgKey(key, false)

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
	registerMsgKey(key, true)

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
//
// The registry is process-global, like the handles themselves: every engine
// validates every declared key against its own locales, so engines sharing a
// process must share a catalog (the common case for a single bot).
var msgRegistry = struct {
	sync.Mutex
	keys map[string]bool // key -> takes template params
}{keys: map[string]bool{}}

func registerMsgKey(key string, hasParams bool) {
	msgRegistry.Lock()
	defer msgRegistry.Unlock()

	msgRegistry.keys[key] = hasParams
}

// resetMsgRegistryForTest clears the registry. Tests only: a deliberately
// missing key would otherwise poison later Validate runs in the package.
func resetMsgRegistryForTest() {
	msgRegistry.Lock()
	defer msgRegistry.Unlock()

	msgRegistry.keys = map[string]bool{}
}

// validateMsgs reports declared message keys (NewMsg/NewMsgP) that the
// default language does not translate. A key whose template fails to render
// without params counts as present — existence is what is validated.
func (l *Languages) validateMsgs() []string {
	msgRegistry.Lock()
	type entry struct {
		key       string
		hasParams bool
	}

	entries := make([]entry, 0, len(msgRegistry.keys))
	for key, hasParams := range msgRegistry.keys {
		entries = append(entries, entry{key, hasParams})
	}
	msgRegistry.Unlock()

	def := l.Default()
	if def == nil {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })

	var missing []string

	for _, e := range entries {
		var err error

		if e.hasParams {
			_, err = def.GetWithParams(e.key, map[string]interface{}{})
		} else {
			_, err = def.Get(e.key)
		}

		var notFound *i18n.MessageNotFoundErr
		if errors.As(err, &notFound) {
			missing = append(missing, e.key)
		}
	}

	return missing
}

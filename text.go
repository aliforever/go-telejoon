package telejoon

import "fmt"

// Text renders a string for a request. Construct values with S (static),
// L (localized), LP (localized with params), D (deferred), K (from session
// storage), or F (fmt.Sprintf composition).
type Text func(ctx *Ctx) string

// S returns a static text.
func S(text string) Text {
	return func(*Ctx) string {
		return text
	}
}

// L returns a text localized by the given message key.
// Falls back to the key itself when languages are not configured or the key
// is not translated; users without a chosen language get the default
// language.
func L(key string) Text {
	return func(ctx *Ctx) string {
		if lang := ctx.resolveLanguage(); lang != nil {
			if text, _ := lang.Get(key); text != "" {
				return text
			}
		}

		return key
	}
}

// LP returns a text localized by the given message key with template params.
// Prefer a declared MsgP handle when the params are known at compile time.
func LP(key string, params map[string]interface{}) Text {
	return func(ctx *Ctx) string {
		if lang := ctx.resolveLanguage(); lang != nil {
			if text, _ := lang.GetWithParams(key, params); text != "" {
				return text
			}
		}

		return key
	}
}

// D returns a deferred text rendered by fn for each request.
func D(fn func(ctx *Ctx) string) Text {
	return fn
}

// K returns a text rendered from the string stored under key in the session.
func K(key Key[string]) Text {
	return func(ctx *Ctx) string {
		return ctx.GetOr(key, "")
	}
}

// F composes texts with fmt.Sprintf.
func F(placeholder string, texts ...Text) Text {
	return func(ctx *Ctx) string {
		args := make([]any, len(texts))
		for i, text := range texts {
			args[i] = text(ctx)
		}

		return fmt.Sprintf(placeholder, args...)
	}
}

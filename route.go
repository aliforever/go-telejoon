package telejoon

import (
	"fmt"
	"net/url"
	"strings"
)

// maxCallbackDataBytes is Telegram's callback_data size limit.
const maxCallbackDataBytes = 64

// Route is a typed callback route. It is created by InlineMenuBuilder.Route
// (menu-scoped) or Engine.Route (global), and bound to buttons with Do:
//
//	del := products.Route("del", func(ctx *telejoon.Ctx, a DelArgs) telejoon.Action { ... })
//	telejoon.Do(telejoon.S("🗑"), del, DelArgs{ProductID: p.ID})
type Route[A any] struct {
	menu  string // empty for engine-global routes
	name  string
	codec Codec[A]
}

// RouteOption configures a Route.
type RouteOption[A any] func(*Route[A])

// WithCodec overrides the route's default positional codec.
func WithCodec[A any](codec Codec[A]) RouteOption[A] {
	return func(route *Route[A]) {
		route.codec = codec
	}
}

// encode renders the callback_data for the given payload, enforcing
// Telegram's 64-byte limit.
func (r Route[A]) encode(args A) (string, error) {
	payload, err := r.codec.Encode(args)
	if err != nil {
		return "", fmt.Errorf("encode_route_%s: %w", r.name, err)
	}

	var data string
	if r.menu != "" {
		data = r.menu + ":" + r.name + ":" + url.QueryEscape(string(payload))
	} else {
		data = r.name + ":" + url.QueryEscape(string(payload))
	}

	if len(data) > maxCallbackDataBytes {
		return "", fmt.Errorf("callback_data_too_long: route %s payload exceeds %d bytes", r.name, maxCallbackDataBytes)
	}

	return data, nil
}

// erasedRouteHandler decodes the wire payload and invokes the typed handler.
type erasedRouteHandler func(ctx *Ctx, escapedPayload string) Action

func (r Route[A]) erase(fn func(ctx *Ctx, args A) Action) erasedRouteHandler {
	return func(ctx *Ctx, escapedPayload string) Action {
		raw, err := url.QueryUnescape(escapedPayload)
		if err != nil {
			return Error(fmt.Errorf("decode_route_%s: %w", r.name, err))
		}

		args, err := r.codec.Decode([]byte(raw))
		if err != nil {
			return Error(fmt.Errorf("decode_route_%s: %w", r.name, err))
		}

		return fn(ctx, args)
	}
}

// validRouteName reports whether name can be used as a route, menu, or state
// identifier in the callback wire protocol.
func validRouteName(name string) bool {
	return name != "" && !strings.Contains(name, ":") && !strings.HasPrefix(name, "@")
}

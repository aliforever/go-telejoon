package telejoon

import "sync/atomic"

// Cond is a reusable, memoized condition. A Cond is evaluated at most once
// per keyboard render: the result is cached in the Ctx, so any number of
// buttons sharing a condition evaluate it once. State switches and
// post-handler re-renders clear the cache, because a handler may have
// changed what the condition reads — a condition is re-evaluated once per
// entered state, not once per request. For a deliberately unshared
// predicate use If(fn); for an expensive check that must run at most once
// per update regardless of switches, cache it yourself with a Key:
//
//	var IsAdmin = telejoon.DefineCond(func(ctx *telejoon.Ctx) bool {
//		return ctx.GetOr(IsAdminKey, false)
//	})
//
//	buttons: telejoon.GoTo(telejoon.L("Nav.Admin"), Admin).When(IsAdmin)
type Cond struct {
	id uint64
	fn func(ctx *Ctx) bool
}

var condCounter atomic.Uint64

// DefineCond creates a new memoized condition.
func DefineCond(fn func(ctx *Ctx) bool) Cond {
	return Cond{id: condCounter.Add(1), fn: fn}
}

// eval returns the condition's value for this request, evaluating and caching
// it on first use.
func (c Cond) eval(ctx *Ctx) bool {
	if ctx.condResults == nil {
		ctx.condResults = map[uint64]bool{}
	}

	if result, ok := ctx.condResults[c.id]; ok {
		return result
	}

	result := c.fn(ctx)
	ctx.condResults[c.id] = result

	return result
}

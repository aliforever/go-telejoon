package main

import (
	"strconv"
	"sync"
)

// === Demo domain data ===

type product struct {
	ID      int64
	Name    string
	Price   float64
	InStock bool
}

var products = []product{
	{ID: 1, Name: "Red Mug", Price: 4.99, InStock: true},
	{ID: 2, Name: "Blue Mug", Price: 5.99, InStock: true},
	{ID: 3, Name: "Golden Mug", Price: 49.99, InStock: false},
}

func productByID(id int64) *product {
	for i := range products {
		if products[i].ID == id {
			return &products[i]
		}
	}

	return nil
}

// === Cross-request demo stores ===
//
// Data that must survive across updates lives in your own storage (here:
// in-memory maps; a real bot would use a database). Ctx session Keys are
// per-request only — storing a cart in one would silently lose it.

var (
	loggedInUsers sync.Map // userID -> true
	carts         sync.Map // userID -> []int64 (product IDs)
	orders        sync.Map // userID -> []order
)

type order struct {
	ProductID int64
	Qty       int
	Address   string
}

// isAdminUser is the demo admin rule: Telegram user 1 is the admin.
// A real bot would check a database.
func isAdminUser(userID int64) bool {
	return userID == 1
}

func isLoggedInUser(userID int64) bool {
	_, ok := loggedInUsers.Load(userID)

	return ok
}

func cartOf(userID int64) []int64 {
	cart, _ := carts.Load(userID)
	ids, _ := cart.([]int64)

	return ids
}

func addToCart(userID int64, productID int64) {
	carts.Store(userID, append(cartOf(userID), productID))
}

func placeOrder(userID int64, o order) {
	list, _ := orders.Load(userID)
	userOrders, _ := list.([]order)

	orders.Store(userID, append(userOrders, o))
}

// === StateDataRepository ===
//
// IMPORTANT: a GoToWith payload rides the transition request itself, but the
// user's NEXT message (e.g. the checkout address) is a new update — the
// payload is then reloaded from this repository. Without
// engine.WithStateDataRepository, the follow-up handlers receive a zero D.
// The engine encodes payloads as JSON.

type memoryStateDataRepository struct {
	data sync.Map // map[stateDataKey][]byte
}

type stateDataKey struct {
	userID int64
	state  string
}

func (r *memoryStateDataRepository) SetUserStateData(userID int64, state string, data []byte) error {
	r.data.Store(stateDataKey{userID, state}, data)

	return nil
}

func (r *memoryStateDataRepository) GetUserStateData(userID int64, state string) ([]byte, error) {
	raw, _ := r.data.Load(stateDataKey{userID, state})
	if raw == nil {
		return nil, nil
	}

	return raw.([]byte), nil
}

// === Custom codec ===
//
// base36Codec is a domain-specific Codec[int64]: order IDs encoded in base36
// are shorter than decimal, buying room under Telegram's 64-byte
// callback_data limit. Register it per route with telejoon.WithCodec.

type base36Codec struct{}

func (base36Codec) Encode(v int64) ([]byte, error) {
	return []byte(strconv.FormatInt(v, 36)), nil
}

func (base36Codec) Decode(data []byte) (int64, error) {
	return strconv.ParseInt(string(data), 36, 64)
}

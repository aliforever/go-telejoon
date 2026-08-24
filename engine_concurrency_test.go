package telejoon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	"golang.org/x/text/language"
)

// === Per-chat serialization through Start ===

// trackingProcessor records per-chat overlap and processing order.
type trackingProcessor struct {
	mu        sync.Mutex
	inFlight  map[int64]int
	overlaps  map[int64]bool
	order     map[int64][]int64
	started1  chan struct{}
	started2  chan struct{}
	blockChat int64
	blockID   int64

	startOnce sync.Once
}

func (p *trackingProcessor) canProcess(update tgbotapi.Update) bool { return true }

func (p *trackingProcessor) Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	chat := update.Chat().Id

	p.mu.Lock()
	p.inFlight[chat]++
	if p.inFlight[chat] > 1 {
		p.overlaps[chat] = true
	}
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.inFlight[chat]--
		p.mu.Unlock()
	}()

	// The designated update blocks until the other chat's update starts,
	// proving cross-chat parallelism.
	if chat == p.blockChat && update.Message != nil && update.Message.Text == "blocker" {
		close(p.started1)
		select {
		case <-p.started2:
		case <-time.After(2 * time.Second):
		}
	}

	if chat != p.blockChat {
		p.startOnce.Do(func() { close(p.started2) })
	}

	p.mu.Lock()
	p.order[chat] = append(p.order[chat], update.UpdateID)
	p.mu.Unlock()
}

func TestStartSerializesPerChatAndParallelizesAcrossChats(t *testing.T) {
	// One batch: a blocking chat-1 update, nine more chat-1 updates, and a
	// chat-2 update — delivered in this order.
	var updates []string

	updates = append(updates,
		`{"update_id":1,"message":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"},"from":{"id":1},"text":"blocker"}}`)

	for i := 2; i <= 10; i++ {
		updates = append(updates, fmt.Sprintf(
			`{"update_id":%d,"message":{"message_id":%d,"date":1,"chat":{"id":1,"type":"private"},"from":{"id":1},"text":"m%d"}}`,
			i, i, i))
	}

	updates = append(updates,
		`{"update_id":11,"message":{"message_id":11,"date":1,"chat":{"id":2,"type":"private"},"from":{"id":2},"text":"other"}}`)

	var served atomic.Bool

	processor := &trackingProcessor{
		inFlight:  map[int64]int{},
		overlaps:  map[int64]bool{},
		order:     map[int64][]int64{},
		started1:  make(chan struct{}),
		started2:  make(chan struct{}),
		blockChat: 1,
	}

	server := newBatchServer(t, &served, updates)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- Start(ctx, server, processor)
	}()

	// Chat 1's blocker must start, and chat 2's update must run while chat
	// 1's blocker is still parked — proving cross-chat parallelism.
	select {
	case <-processor.started1:
	case <-time.After(2 * time.Second):
		t.Fatal("chat 1 blocker never started")
	}

	select {
	case <-processor.started2:
	case <-time.After(2 * time.Second):
		t.Fatal("chat 2 update did not run while chat 1 was blocked — chats are not parallel")
	}

	// Wait until every queued update has been processed, then stop.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		processor.mu.Lock()
		processed := len(processor.order[1]) + len(processor.order[2])
		processor.mu.Unlock()

		if processed == 11 {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	cancel()

	if err := <-done; err != context.Canceled {
		t.Fatalf("Start err = %v, want context.Canceled", err)
	}

	if processor.overlaps[1] || processor.overlaps[2] {
		t.Fatalf("same-chat updates overlapped: %v", processor.overlaps)
	}

	// Chat 1 processed its updates in arrival order.
	want := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	processor.mu.Lock()
	got := processor.order[1]
	processor.mu.Unlock()

	if len(got) != len(want) {
		t.Fatalf("chat 1 processed %d updates, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("chat 1 order = %v, want %v", got, want)
		}
	}
}

// newBatchServer returns a bot client whose test server serves the given raw
// updates once, then empty getUpdates batches.
func newBatchServer(t *testing.T, served *atomic.Bool, updates []string) *tgbotapi.Bot {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path.Base(r.URL.Path) {
		case "getUpdates":
			if served != nil && !served.Swap(true) {
				fmt.Fprintf(w, `{"ok":true,"result":[%s]}`, strings.Join(updates, ","))

				return
			}

			fmt.Fprint(w, `{"ok":true,"result":[]}`)
		case "sendMessage":
			fmt.Fprint(w, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"},"text":"ok"}}`)
		default:
			fmt.Fprint(w, `{"ok":true,"result":true}`)
		}
	}))

	t.Cleanup(server.Close)

	return tgbotapi.New("test-token", tgbotapi.WithAPIURL(server.URL))
}

// === Engine-level concurrency hammer ===

// TestConcurrentProcessIsRaceFree hammers Engine.Process from many
// goroutines — same user and many users, messages and callbacks — relying on
// the race detector and on post-hoc state consistency.
func TestConcurrentProcessIsRaceFree(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	repo := NewMockUserRepository()
	dataRepo := NewDefaultStateDataRepository()

	stateBuy := NewState[purchaseData]("ConcBuy")

	engine := New(repo, stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, NewMockUserLanguageRepository())).
		WithStateDataRepository(dataRepo).
		Use(func(ctx *Ctx) Action { return Next() })

	menuRef := NewInlineMenuRef("ConcCatalog")

	catalog := InlineMenuFor(menuRef, S("catalog"))
	buy := catalog.Route("buy", func(ctx *Ctx, args purchaseData) Action {
		return ctx.GoToWith(stateBuy, purchaseData{ProductID: args.ProductID})
	})
	catalog.Buttons(Do(S("buy"), buy, purchaseData{ProductID: 1}))

	engine.Add(
		Menu(stateWelcome, S("welcome")).
			Buttons(ShowInline(S("catalog"), menuRef)).
			OnText(func(ctx *Ctx, _ *NoData, text string) Action {
				return ctx.ReplyText("echo: " + text)
			}),
		Menu(stateBuy, S("buying")).
			OnText(func(ctx *Ctx, data *purchaseData, text string) Action {
				return ctx.GoTo(stateWelcome)
			}),
		catalog,
	)

	ts, bot := newTelegramServer(t)

	const users = 20
	const updatesPerUser = 10

	var wg sync.WaitGroup

	for user := 0; user < users; user++ {
		wg.Add(1)

		go func(user int64) {
			defer wg.Done()

			for i := 0; i < updatesPerUser; i++ {
				var update tgbotapi.Update

				if i%3 == 2 {
					// Positional codec: "1:0" double-escaped for the wire.
					update = NewTestUpdate().WithCallbackQuery(user, "ConcCatalog:buy:1%3A0").Build()
				} else {
					update = NewTestUpdate().WithMessage(user, user, "msg "+strconv.Itoa(i)).Build()
				}

				engine.Process(context.Background(), bot, update)
			}
		}(int64(user))
	}

	// Plus one user hammered by parallel goroutines (unsynchronized at this
	// layer — Start's serializer is what orders these in production; here we
	// only demand race-freedom and a valid end state).
	for g := 0; g < 8; g++ {
		wg.Add(1)

		go func(g int) {
			defer wg.Done()

			for i := 0; i < 10; i++ {
				engine.Process(context.Background(), bot,
					NewTestUpdate().WithMessage(999, 999, fmt.Sprintf("g%d-%d", g, i)).Build())
			}
		}(g)
	}

	wg.Wait()

	// Every user must end in a registered state.
	for user := int64(0); user < users; user++ {
		state, err := repo.GetUserState(user)
		if err != nil {
			t.Fatalf("user %d has no state: %v", user, err)
		}

		if engine.getMenu(state) == nil {
			t.Fatalf("user %d ended in unregistered state %q", user, state)
		}
	}

	// Every processed update produces at least one API call (a render or an
	// echo): 20 users × 10 updates + 80 hammer messages = 280.
	ts.waitFor(t, "all sends", func() bool {
		return len(ts.callsOf("sendMessage")) >= 280
	})
}

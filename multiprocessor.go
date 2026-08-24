package telejoon

import (
	"context"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// MultiProcessor routes updates to multiple processors based on their canProcess method.
// The first processor that can handle the update will process it.
type MultiProcessor struct {
	mu         sync.RWMutex
	processors []Processor
}

// NewMultiProcessor creates a processor that routes to multiple sub-processors.
// Processors are checked in order; the first one that can process the update handles it.
//
// Example:
//
//	engine := telejoon.New(userRepo, stateWelcome)
//	groupHandler := telejoon.NewGroupHandlers()
//	channelHandler := telejoon.NewChannelHandlers()
//
//	multi := telejoon.NewMultiProcessor(engine, groupHandler, channelHandler)
//	telejoon.Start(ctx, client, multi)
func NewMultiProcessor(processors ...Processor) *MultiProcessor {
	return &MultiProcessor{
		processors: processors,
	}
}

// AddProcessor adds a processor to the multi-processor.
// Processors are checked in the order they are added.
func (m *MultiProcessor) AddProcessor(p Processor) *MultiProcessor {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processors = append(m.processors, p)
	return m
}

func (m *MultiProcessor) snapshot() []Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]Processor(nil), m.processors...)
}

func (m *MultiProcessor) canProcess(update tgbotapi.Update) bool {
	for _, p := range m.snapshot() {
		if p.canProcess(update) {
			return true
		}
	}
	return false
}

func (m *MultiProcessor) Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	for _, p := range m.snapshot() {
		if p.canProcess(update) {
			p.Process(ctx, client, update)
			return
		}
	}
}

// ProcessAll processes the update with all processors that can handle it.
// Unlike Process, this doesn't stop after the first matching processor.
func (m *MultiProcessor) ProcessAll(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	for _, p := range m.snapshot() {
		if p.canProcess(update) {
			p.Process(ctx, client, update)
		}
	}
}

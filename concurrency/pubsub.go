package concurrency

import "sync"

// PubSub implements a concurrency-safe publisher/subscriber model.
type PubSub[T any] struct {
	sync.Mutex

	id   uint
	subs map[uint]chan T
}

// NewPubSub creates a new empty pubsub.
func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		id:   0,
		subs: make(map[uint]chan T),
	}
}

// Publish publishes the given message to all subscribers.
func (p *PubSub[T]) Publish(message T) {
	p.Lock()
	defer p.Unlock()
	for _, dst := range p.subs {
		select {
		case dst <- message:
		default:
		}
	}
}

// Subscribe subscribes to all messages, and returns a function that can be used
// to close the subscription.
func (p *PubSub[T]) Subscribe() (<-chan T, func()) {
	p.Lock()
	defer p.Unlock()
	p.id += 1
	p.subs[p.id] = make(chan T)
	return p.subs[p.id], func() {
		p.Lock()
		defer p.Unlock()
		delete(p.subs, p.id)
	}
}

// SubscribeBuffered behaves the same as `Subscribe` but allows configuring a
// buffer size to help avoid dropped messages.
func (p *PubSub[T]) SubscribeBuffered(bufferSize int) (<-chan T, func()) {
	p.Lock()
	defer p.Unlock()
	p.id += 1
	p.subs[p.id] = make(chan T, bufferSize)
	return p.subs[p.id], func() {
		p.Lock()
		defer p.Unlock()
		delete(p.subs, p.id)
	}
}

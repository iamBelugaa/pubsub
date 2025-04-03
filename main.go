package pubsub

import (
	"sync"
)

type PubSub struct {
	closed        bool
	mu            sync.RWMutex
	subscriptions map[string][]chan string
}

func NewPubSub() *PubSub {
	return &PubSub{
		mu:            sync.RWMutex{},
		subscriptions: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan string, 1)
	channels := ps.subscriptions[topic]

	ps.subscriptions[topic] = append(channels, ch)
	return ch
}

func (ps *PubSub) Publish(topic, msg string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}

	channels := ps.subscriptions[topic]
	for _, ch := range channels {
		ch <- msg
	}
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return
	}

	ps.closed = true

	for _, channels := range ps.subscriptions {
		for _, ch := range channels {
			close(ch)
		}
	}
}

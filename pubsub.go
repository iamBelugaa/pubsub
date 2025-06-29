package pubsub

import (
	"errors"
	"sync"
)

var (
	ErrPubSubClosed = errors.New("pubsub is closed")
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

func (ps *PubSub) Subscribe(topic string) (<-chan string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, ErrPubSubClosed
	}

	ch := make(chan string, 1)
	channels := ps.subscriptions[topic]

	ps.subscriptions[topic] = append(channels, ch)
	return ch, nil
}

func (ps *PubSub) Publish(topic, msg string) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	channels := ps.subscriptions[topic]
	for _, ch := range channels {
		ch <- msg
	}

	return nil
}

func (ps *PubSub) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	ps.closed = true

	for _, channels := range ps.subscriptions {
		for _, ch := range channels {
			close(ch)
		}
	}

	return nil
}

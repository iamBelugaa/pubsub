package pubsub

import (
	"errors"
	"sync"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrPubSubClosed  = errors.New("pubsub is closed")
)

type Message struct {
	Topic   string
	Message string
}

type PubSub struct {
	closed        bool
	mu            sync.RWMutex
	subscriptions map[string][]chan *Message
}

func NewPubSub() *PubSub {
	return &PubSub{
		mu:            sync.RWMutex{},
		subscriptions: map[string][]chan *Message{},
	}
}

func (ps *PubSub) Subscribe(topic string) (<-chan *Message, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, ErrPubSubClosed
	}

	subscriber := make(chan *Message, 1)
	subscribers := ps.subscriptions[topic]

	ps.subscriptions[topic] = append(subscribers, subscriber)
	return subscriber, nil
}

func (ps *PubSub) Publish(topic, msg string) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	channels, ok := ps.subscriptions[topic]
	if !ok {
		return ErrTopicNotFound
	}

	for _, ch := range channels {
		ch <- &Message{Topic: topic, Message: msg}
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

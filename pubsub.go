package pubsub

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrTopicNotFound is returned when attempting to publish to a topic
	// that has no subscribers.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrPubSubClosed is returned when performing operations on a closed
	// PubSub system.
	ErrPubSubClosed = errors.New("pubsub is closed")
)

// defaultConfig defines the default configuration for the PubSub system.
// By default, each subscriber channel will have a buffer size of 1.
var defaultConfig = config{
	channelSize: 1,
}

// New initializes a new PubSub system with optional configuration settings.
// Users can pass functional options to modify the default behavior.
func New[T any](options ...Option) *PubSub[T] {
	config := defaultConfig

	// Apply provided configuration options.
	for _, opt := range options {
		opt(&config)
	}

	return &PubSub[T]{
		config:        &config,
		mu:            sync.RWMutex{},
		subscriptions: map[string][]chan *Message[T]{},
	}
}

// Subscribe registers a new subscriber to the specified topic.
// It returns a read-only channel from which the subscriber can receive messages.
//
// Returns an error if the PubSub system is closed.
func (ps *PubSub[T]) Subscribe(topic string) (<-chan *Message[T], error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, ErrPubSubClosed
	}

	// Create a new subscriber channel with the configured buffer size.
	subscriber := make(chan *Message[T], ps.config.channelSize)

	// Append the new subscriber to the list of subscribers for the topic.
	ps.subscriptions[topic] = append(ps.subscriptions[topic], subscriber)
	return subscriber, nil
}

// Publish sends a message to all subscribers of the given topic.
//
// Returns an error if the PubSub system is closed or if there are no subscribers for the topic.
func (ps *PubSub[T]) Publish(topic string, msg T) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	channels, ok := ps.subscriptions[topic]
	if !ok {
		return ErrTopicNotFound
	}

	// Deliver the message to all subscribers of the topic.
	for _, ch := range channels {
		message := Message[T]{Topic: topic, Payload: msg}
		select {
		case ch <- &message:
		default:
			fmt.Printf("Failed to send message for Topic : %s - %#v\n", topic, message)
		}
	}

	return nil
}

// Close shuts down the PubSub system, preventing further publishing and subscribing.
// All open subscription channels are closed.
//
// Returns an error if the system is already closed.
func (ps *PubSub[T]) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	ps.closed = true

	// Close all subscriber channels.
	for _, channels := range ps.subscriptions {
		for _, ch := range channels {
			close(ch)
		}
	}

	return nil
}

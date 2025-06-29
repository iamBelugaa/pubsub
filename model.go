package pubsub

import "sync"

// Message represents a published message in the PubSub system.
type Message[T any] struct {
	Payload T
	Topic   string
}

// PubSub implements a basic publish-subscribe system.
type PubSub[T any] struct {
	closed        bool                          // Indicates if the PubSub system is closed.
	config        *config                       // Configuration settings for the PubSub system.
	mu            sync.RWMutex                  // Mutex to synchronize access to subscriptions.
	subscriptions map[string][]chan *Message[T] // Map of topic subscriptions to channels.
}

// config holds the configuration settings for the PubSub system.
type config struct {
	channelSize int // Defines the buffer size for subscriber channels.
}

type Option func(*config)

// WithChannelSize is used to set the size of the subscriber channels.
func WithChannelSize(size int) Option {
	return func(c *config) {
		if size > 0 {
			c.channelSize = size
		}
	}
}

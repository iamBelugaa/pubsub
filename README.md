# Go PubSub

A lightweight, in-memory Publish-Subscribe (PubSub) system written in Go
leveraging Go's powerful concurrency model with channels.

## Features

- Generic support for any payload type.
- Support for multiple subscribers per topic.

## Installation

```sh
go get github.com/iamBelugaa/pubsub
```

## Implementation Details

### Message Struct

The `Message` struct provides a structured way to receive messages with context:

```go
type Message[T any] struct {
	Topic   string // The topic this message belongs to.
	Payload T      // The actual message payload.
}
```

This allows subscribers to filter or route messages based on topic information,
even when listening to multiple topics. The generic type parameter `T` enables
you to use any type as the payload.

### PubSub Methods

#### `New[T any]()`

Creates a new PubSub instance with proper initialization for the specified
payload type.

#### `Subscribe(topic string) (<-chan *Message[T], error)`

Subscribes to a topic and returns a channel that will receive messages published
to that topic.

#### `Publish(topic string, msg T) error`

Publishes a message to the specified topic. Returns an error if the topic
doesn't exist or if the PubSub system is closed.

#### `Close() error`

Closes the PubSub system, shutting down all subscription channels.

### Configuration

You can configure the PubSub system using options:

```go
ps := pubsub.New[string](pubsub.WithChannelSize(10))
```

This sets the buffer size of subscriber channels to 10 and creates a PubSub
instance that works with string payloads.

## Usage Examples

### Basic Example

This example demonstrates the core functionality of subscribing to topics,
publishing messages, and handling graceful shutdown:

```go
package main

import (
	"fmt"
	"sync"

	"github.com/iamBelugaa/pubsub"
)

func main() {
	// Create a PubSub instance that works with string payloads
	ps := pubsub.New[string](pubsub.WithChannelSize(5))
	var wg sync.WaitGroup

	// Subscribe to different topics
	devopsChan, err := ps.Subscribe("devops")
	if err != nil {
		fmt.Println("Error subscribing to devops:", err)
		return
	}

	golangChan, err := ps.Subscribe("golang")
	if err != nil {
		fmt.Println("Error subscribing to golang:", err)
		return
	}

	// Create goroutines to handle messages
	wg.Add(2)

	go func() {
		defer wg.Done()
		for msg := range devopsChan {
			fmt.Printf("[%s]: %s\n", msg.Topic, msg.Payload)
		}
		fmt.Println("DevOps channel closed")
	}()

	go func() {
		defer wg.Done()
		for msg := range golangChan {
			fmt.Printf("[%s]: %s\n", msg.Topic, msg.Payload)
		}
		fmt.Println("Golang channel closed")
	}()

	// Publish messages to topics
	if err := ps.Publish("golang", "Go is great for concurrency!"); err != nil {
		fmt.Println("Error publishing to golang:", err)
	}
	if err := ps.Publish("devops", "CI/CD pipelines automate deployments."); err != nil {
		fmt.Println("Error publishing to devops:", err)
	}

	// Close PubSub system
	if err := ps.Close(); err != nil {
		fmt.Println("Error closing pubsub:", err)
	}

	wg.Wait()
}
```

### Using Custom Types

This example demonstrates using a custom struct as the payload type:

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/iamBelugaa/pubsub"
)

// Define a custom message type
type EventData struct {
	Timestamp time.Time
	Severity  string
	Content   string
}

func main() {
	// Create a PubSub instance for EventData type
	ps := pubsub.New[EventData]()
	var wg sync.WaitGroup

	// Subscribe to events topic
	eventsChan, err := ps.Subscribe("events")
	if err != nil {
		fmt.Println("Error subscribing to events:", err)
		return
	}

	// Process received events
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range eventsChan {
			evt := event.Payload
			fmt.Printf("[%s] %s: %s (at %s)\n",
				event.Topic,
				evt.Severity,
				evt.Content,
				evt.Timestamp.Format(time.RFC3339),
			)
		}
		fmt.Println("Events channel closed")
	}()

	// Publish events
	ps.Publish("events", EventData{
		Timestamp: time.Now(),
		Severity:  "INFO",
		Content:   "System startup completed",
	})

	ps.Publish("events", EventData{
		Timestamp: time.Now(),
		Severity:  "WARNING",
		Content:   "High memory usage detected",
	})

	// Close PubSub system after a short delay
	time.Sleep(time.Second)
	ps.Close()
	wg.Wait()
}
```

### HTTP Server Example

This example demonstrates how to use the PubSub system within an HTTP server to
push real-time updates:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/iamBelugaa/pubsub"
)

// Define a message structure for updates
type UpdateMessage struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	ps := pubsub.New[UpdateMessage]()
	var wg sync.WaitGroup

	// Handler function that subscribes clients to the "updates" topic
	http.HandleFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		// Set headers for server-sent events
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Subscribe to updates topic
		ch, err := ps.Subscribe("updates")
		if err != nil {
			http.Error(w, "Subscription error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Detect client disconnection
		notify := r.Context().Done()
		go func() {
			<-notify
			fmt.Println("Client disconnected")
		}()

		// Stream messages as they arrive
		for msg := range ch {
			// Create a response object
			response := struct {
				Topic     string       `json:"topic"`
				Content   string       `json:"content"`
				Timestamp time.Time    `json:"timestamp"`
			}{
				Topic:     msg.Topic,
				Content:   msg.Payload.Content,
				Timestamp: msg.Payload.Timestamp,
			}

			// Convert to JSON
			eventData, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			w.(http.Flusher).Flush()
		}
	})

	// Start HTTP server
	go func() {
		fmt.Println("Server started at http://localhost:8080/updates")
		http.ListenAndServe(":8080", nil)
	}()

	// Simulate a background process publishing updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			ps.Publish("updates", UpdateMessage{
				Content:   fmt.Sprintf("System update %d: Processing completed", i),
				Timestamp: time.Now(),
			})
			time.Sleep(2 * time.Second)
		}
		ps.Close()
	}()

	wg.Wait()
}
```

## Error Handling

The library provides specific error types to handle different scenarios:

- `ErrTopicNotFound`: Returned when attempting to publish to a non-existent
  topic
- `ErrPubSubClosed`: Returned when attempting operations on a closed PubSub
  instance

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

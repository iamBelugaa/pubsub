# Go PubSub

A lightweight, in-memory Publish-Subscribe (PubSub) system written in Go
leveraging Go's powerful concurrency model with channels. This library enables
multiple subscribers to listen to topics and receive messages asynchronously.

## Features

- Simple and efficient PubSub implementation built on Go's concurrency
  primitives.
- Support for multiple subscribers per topic.
- Structured message delivery with topic information included.
- Thread-safe operations with proper locking mechanisms.
- Graceful shutdown capabilities for safe channel closure.
- Generic support for any payload type.

## Installation

```sh
go get github.com/iamNilotpal/pubsub
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

	"github.com/iamNilotpal/pubsub"
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

	"github.com/iamNilotpal/pubsub"
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

	"github.com/iamNilotpal/pubsub"
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

### Complex Example: Multi-Topic Monitoring System

This example showcases a more comprehensive implementation with multiple topics
and different message types:

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/iamNilotpal/pubsub"
)

// Define different event types
type BaseEvent struct {
	Timestamp time.Time
	Source    string
}

type SystemEvent struct {
	BaseEvent
	Action string
	Status string
}

type SecurityEvent struct {
	BaseEvent
	IPAddress string
	Username  string
	Action    string
}

type PerformanceMetric struct {
	BaseEvent
	Metric string
	Value  float64
	Unit   string
}

type ErrorEvent struct {
	BaseEvent
	Code        int
	Description string
	Severity    string
}

type Event any

func main() {
	// Create PubSub with Event to handle different event types
	ps := pubsub.New[Event]()
	var wg sync.WaitGroup

	// Define topics
	topics := []string{"system", "security", "performance", "errors"}

	// Subscribe to all topics
	topicChannels := make(map[string]<-chan *pubsub.Message[Event])

	for _, topic := range topics {
		ch, err := ps.Subscribe(topic)
		if err != nil {
			fmt.Printf("Failed to subscribe to %s: %v\n", topic, err)
			continue
		}
		topicChannels[topic] = ch
	}

	// Process messages from each topic
	for topic, ch := range topicChannels {
		wg.Add(1)
		go func(topic string, msgChan <-chan *pubsub.Message[Event]) {
			defer wg.Done()

			for msg := range msgChan {
				timestamp := time.Now().Format("15:04:05")

				switch topic {
				case "system":
					if evt, ok := msg.Payload.(SystemEvent); ok {
						fmt.Printf("🖥️ [%s] SYSTEM: %s - %s (%s)\n",
							timestamp, evt.Action, evt.Status, evt.Source)
					}

				case "security":
					if evt, ok := msg.Payload.(SecurityEvent); ok {
						fmt.Printf("🔒 [%s] SECURITY: %s by %s from %s (%s)\n",
							timestamp, evt.Action, evt.Username, evt.IPAddress, evt.Source)
					}

				case "performance":
					if evt, ok := msg.Payload.(PerformanceMetric); ok {
						fmt.Printf("📊 [%s] METRIC: %s = %.2f%s (%s)\n",
							timestamp, evt.Metric, evt.Value, evt.Unit, evt.Source)
					}

				case "errors":
					if evt, ok := msg.Payload.(ErrorEvent); ok {
						fmt.Printf("❌ [%s] ERROR[%d]: %s - %s (%s)\n",
							timestamp, evt.Code, evt.Description, evt.Severity, evt.Source)
					}
				}
			}

			fmt.Printf("Channel for topic '%s' closed\n", topic)
		}(topic, ch)
	}

	// Publish different types of events
	now := time.Now()

	// System events
	ps.Publish("system", SystemEvent{
		BaseEvent: BaseEvent{Timestamp: now, Source: "kernel"},
		Action:    "System startup",
		Status:    "completed",
	})

	ps.Publish("system", SystemEvent{
		BaseEvent: BaseEvent{Timestamp: now, Source: "scheduler"},
		Action:    "Job processing",
		Status:    "in progress",
	})

	// Security events
	ps.Publish("security", SecurityEvent{
		BaseEvent: BaseEvent{Timestamp: now, Source: "auth-service"},
		IPAddress: "192.168.1.254",
		Username:  "admin",
		Action:    "Login attempt",
	})

	ps.Publish("security", SecurityEvent{
		BaseEvent: BaseEvent{Timestamp: now, Source: "user-service"},
		IPAddress: "10.0.0.1",
		Username:  "system",
		Action:    "Permission change",
	})

	// Performance metrics
	ps.Publish("performance", PerformanceMetric{
		BaseEvent: BaseEvent{Timestamp: now, Source: "monitor-service"},
		Metric:    "CPU Usage",
		Value:     65.5,
		Unit:      "%",
	})

	ps.Publish("performance", PerformanceMetric{
		BaseEvent: BaseEvent{Timestamp: now, Source: "monitor-service"},
		Metric:    "Memory",
		Value:     2.45,
		Unit:      "GB",
	})

	// Error events
	ps.Publish("errors", ErrorEvent{
		BaseEvent:   BaseEvent{Timestamp: now, Source: "database"},
		Code:        5001,
		Description: "Connection timeout",
		Severity:    "high",
	})

	ps.Publish("errors", ErrorEvent{
		BaseEvent:   BaseEvent{Timestamp: now, Source: "api-gateway"},
		Code:        4029,
		Description: "Rate limit exceeded",
		Severity:    "medium",
	})

	// Wait briefly to ensure message processing
	time.Sleep(time.Second)

	// Close the PubSub system
	if err := ps.Close(); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("All subscribers have terminated successfully")
}
```

## Error Handling

The library provides specific error types to handle different scenarios:

- `ErrTopicNotFound`: Returned when attempting to publish to a non-existent
  topic
- `ErrPubSubClosed`: Returned when attempting operations on a closed PubSub
  instance

## Best Practices

- Create a new PubSub instance for each logical separation of concerns.
- Always check for errors when subscribing or publishing.
- Use `defer` to ensure proper closure of the PubSub system.
- Implement proper context handling for HTTP-based implementations.
- Consider using channel buffering for high-throughput scenarios.
- Use a termination signal (like a `done` channel) to cleanly shut down
  goroutines.
- Choose appropriate generic types based on your use case:
  - Use specific types like `string` or custom structs for type safety.
  - Use `interface{}` or `any` when flexibility is needed.

## Thread Safety

All operations on the PubSub instance are thread-safe, utilizing a read-write
mutex to coordinate access to the subscription map.

## Performance Considerations

The PubSub system is designed for in-memory operations within a single process.
For distributed systems requiring cross-service communication, consider using
specialized message brokers like RabbitMQ, Kafka, or NATS.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

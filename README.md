# Go PubSub

A lightweight, in-memory Publish-Subscribe (PubSub) system written in Go
leveraging Go's powerful concurrency model with channels. This library enables
multiple subscribers to listen to topics and receive messages asynchronously.

## Features

- Simple and efficient PubSub implementation built on Go's concurrency
  primitives
- Support for multiple subscribers per topic
- Structured message delivery with topic information included
- Thread-safe operations with proper locking mechanisms
- Graceful shutdown capabilities for safe channel closure
- Easy integration into web applications and microservices

## Installation

```sh
go get github.com/iamNilotpal/pubsub
```

## Key Concepts

### Message Structure

Messages in the PubSub system are delivered as `Message` structs:

```go
type Message struct {
  Topic   string // The topic this message was published to
  Message string // The actual message content
}
```

This structure allows subscribers to know which topic a message originated from,
enabling more sophisticated message handling.

### Configuration

You can configure the PubSub system using options:

```go
ps := pubsub.NewPubSub(pubsub.WithChannelSize(10))
```

This sets the buffer size of subscriber channels to 10.

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
	ps := pubsub.NewPubSub(pubsub.WithChannelSize(5))
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
			fmt.Printf("[%s]: %s\n", msg.Topic, msg.Message)
		}
		fmt.Println("DevOps channel closed")
	}()

	go func() {
		defer wg.Done()
		for msg := range golangChan {
			fmt.Printf("[%s]: %s\n", msg.Topic, msg.Message)
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

func main() {
	ps := pubsub.NewPubSub()
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
			// Convert message to JSON
			eventData, _ := json.Marshal(map[string]string{
				"topic":   msg.Topic,
				"message": msg.Message,
				"time":    time.Now().Format(time.RFC3339),
			})

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
			ps.Publish("updates", fmt.Sprintf("System update %d: Processing completed", i))
			time.Sleep(2 * time.Second)
		}
		ps.Close()
	}()

	wg.Wait()
}
```

### Complex Example: Multi-Topic Monitoring System

This example showcases a more comprehensive implementation with multiple topics,
error handling, and the use of the Message struct for advanced filtering:

```go
package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iamNilotpal/pubsub"
)

func main() {
	ps := pubsub.NewPubSub()
	var wg sync.WaitGroup

	// Define topics
	topics := []string{"system", "security", "performance", "errors"}

	// Create a channel to collect all messages
	// Use a buffered channel to avoid deadlocks
	allMessages := make(chan *pubsub.Message, 100)

	// Variable to track when we should stop processing
	done := make(chan struct{})

	// Subscribe to all topics and set up forwarders
	for _, topic := range topics {
		topicCh, err := ps.Subscribe(topic)
		if err != nil {
			fmt.Printf("Failed to subscribe to %s: %v\n", topic, err)
			continue
		}

		// For each subscription, forward messages to the collector channel
		go func(ch <-chan *pubsub.Message) {
			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						// Channel closed, exit goroutine
						return
					}
					// Forward the message
					allMessages <- msg
				case <-done:
					// Exit signal received
					return
				}
			}
		}(topicCh)
	}

	// Message processor goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case msg, ok := <-allMessages:
				if !ok {
					// Channel closed, exit goroutine
					return
				}

				timestamp := time.Now().Format("15:04:05")

				// Handle messages differently based on topic
				switch msg.Topic {
				case "security":
					fmt.Printf("🔒 [%s] SECURITY ALERT: %s\n", timestamp, msg.Message)

					// Apply additional security-specific logic
					if strings.Contains(msg.Message, "Failed login") {
						fmt.Println("    ⚠️  Potential breach attempt detected!")
					}

				case "performance":
					fmt.Printf("📊 [%s] PERFORMANCE: %s\n", timestamp, msg.Message)

					// Extract metrics if present
					if strings.Contains(msg.Message, "CPU usage") {
						parts := strings.Split(msg.Message, ":")
						if len(parts) > 1 {
							fmt.Printf("    System load detected at%s\n", parts[1])
						}
					}

				case "errors":
					fmt.Printf("❌ [%s] ERROR: %s\n", timestamp, msg.Message)

				default:
					fmt.Printf("ℹ️ [%s] [%s] %s\n", timestamp, msg.Topic, msg.Message)
				}

			case <-done:
				// Exit signal received
				fmt.Println("Message processor shutting down")
				return
			}
		}
	}()

	// Publish messages with different patterns
	go func() {
		// System messages - periodic updates
		for i := 1; i <= 3; i++ {
			ps.Publish("system", fmt.Sprintf("System check %d completed", i))
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		// Security alerts - random patterns
		alerts := []string{
			"Failed login attempt detected from IP 192.168.1.254",
			"Configuration file accessed by user admin",
			"New user account created: operator",
		}

		for _, alert := range alerts {
			ps.Publish("security", alert)
			time.Sleep(1500 * time.Millisecond)
		}
	}()

	go func() {
		// Performance metrics
		metrics := map[string]int{
			"CPU usage":     65,
			"Memory usage":  78,
			"Disk I/O":      45,
			"Network usage": 30,
		}

		for metric, value := range metrics {
			ps.Publish("performance", fmt.Sprintf("%s: %d%%", metric, value))
			time.Sleep(800 * time.Millisecond)
		}
	}()

	go func() {
		// Simulate error conditions
		time.Sleep(3 * time.Second)
		ps.Publish("errors", "Database connection timeout")
		time.Sleep(1 * time.Second)
		ps.Publish("errors", "API rate limit exceeded")
	}()

	// Run for a while then close gracefully
	time.Sleep(6 * time.Second)
	fmt.Println("\nInitiating graceful shutdown...")

	// Signal all goroutines to stop
	close(done)

	// Close the PubSub system
	if err := ps.Close(); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}

	// Close the collector channel after PubSub is closed
	close(allMessages)

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

## Implementation Details

### Message Struct

The `Message` struct provides a structured way to receive messages with context:

```go
type Message struct {
	Topic   string // The topic this message belongs to
	Message string // The actual message content
}
```

This allows subscribers to filter or route messages based on topic information,
even when listening to multiple topics.

### PubSub Methods

#### NewPubSub()

Creates a new PubSub instance with proper initialization.

#### Subscribe(topic string) (<-chan \*Message, error)

Subscribes to a topic and returns a channel that will receive messages published
to that topic.

#### Publish(topic, msg string) error

Publishes a message to the specified topic. Returns an error if the topic
doesn't exist or if the PubSub system is closed.

#### Close() error

Closes the PubSub system, shutting down all subscription channels.

## Best Practices

- Create a new PubSub instance for each logical separation of concerns
- Always check for errors when subscribing or publishing
- Use `defer` to ensure proper closure of the PubSub system
- Implement proper context handling for HTTP-based implementations
- Consider using channel buffering for high-throughput scenarios
- Use a termination signal (like a `done` channel) to cleanly shut down
  goroutines

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

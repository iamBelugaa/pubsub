# 📢 Go PubSub

A lightweight, in-memory **Publish-Subscribe (PubSub) System** written in Go
using **channels**. This library allows multiple subscribers to listen to topics
and receive messages asynchronously.

## 🚀 Features

- Simple and efficient **PubSub** implementation using Go's concurrency model.
- Supports **multiple subscribers** per topic.
- **Graceful shutdown** to close all channels safely.
- Easy integration into **web applications** and **microservices**.

## 📦 Installation

```sh
go get "github.com/iamNilotpal/pubsub/pubsub"
```

## 📖 Usage

### **1️⃣ Basic Example**

This example demonstrates subscribing to topics, publishing messages, and
handling graceful shutdown.

```go
package main

import (
	"fmt"
	"sync"
	"github.com/iamNilotpal/pubsub/pubsub"
)

func main() {
	pubsub := pubsub.NewPubSub()
	var wg sync.WaitGroup

	devopsChan, err := pubsub.Subscribe("devops")
	if err != nil {
		fmt.Println("Error subscribing to devops:", err)
		return
	}

golangChan, err := pubsub.Subscribe("golang")
	if err != nil {
		fmt.Println("Error subscribing to golang:", err)
		return
	}

kubernetesChan, err := pubsub.Subscribe("kubernetes")
	if err != nil {
		fmt.Println("Error subscribing to kubernetes:", err)
		return
	}

	wg.Add(3)

	// Goroutine to listen for messages on "devops" topic
	go func() {
		defer wg.Done()
		for msg := range devopsChan {
			fmt.Println("[DevOps]:", msg)
		}
		fmt.Println("DevOps channel closed")
	}()

	// Goroutine to listen for messages on "golang" topic
	go func() {
		defer wg.Done()
		for msg := range golangChan {
			fmt.Println("[Golang]:", msg)
		}
		fmt.Println("Golang channel closed")
	}()

	// Goroutine to listen for messages on "kubernetes" topic
	go func() {
		defer wg.Done()
		for msg := range kubernetesChan {
			fmt.Println("[Kubernetes]:", msg)
		}
		fmt.Println("Kubernetes channel closed")
	}()

	// Publish messages
	if err := pubsub.Publish("golang", "Go is great for concurrency!"); err != nil {
		fmt.Println("Error publishing to golang:", err)
	}
	if err := pubsub.Publish("devops", "CI/CD pipelines automate deployments."); err != nil {
		fmt.Println("Error publishing to devops:", err)
	}
	if err := pubsub.Publish("kubernetes", "K8s makes container orchestration easy."); err != nil {
		fmt.Println("Error publishing to kubernetes:", err)
	}

	// Close PubSub
	if err := pubsub.Close(); err != nil {
		fmt.Println("Error closing pubsub:", err)
	}

	wg.Wait()
}
```

### **2️⃣ Multiple Subscribers to the Same Topic**

This example shows how multiple subscribers can listen to the same topic.

```go
package main

import (
	"fmt"
	"sync"
	ps "github.com/iamNilotpal/pubsub/pubsub"
)

func main() {
	ps := pubsub.NewPubSub()
	var wg sync.WaitGroup

	// Subscribe multiple listeners to "news" topic
	newsChan1, err := ps.Subscribe("news")
	if err != nil {
		fmt.Println("Error subscribing to news:", err)
		return
	}
	newsChan2, err := ps.Subscribe("news")
	if err != nil {
		fmt.Println("Error subscribing to news:", err)
		return
	}

	wg.Add(2)

	// First subscriber
	go func() {
		defer wg.Done()
		for msg := range newsChan1 {
			fmt.Println("[Subscriber 1]:", msg)
		}
	}()

	// Second subscriber
	go func() {
		defer wg.Done()
		for msg := range newsChan2 {
			fmt.Println("[Subscriber 2]:", msg)
		}
	}()

	// Publisher sends messages
	if err := ps.Publish("news", "Breaking News: Golang 1.22 released!"); err != nil {
		fmt.Println("Error publishing to news:", err)
	}
	if err := ps.Publish("news", "New Kubernetes security update available."); err != nil {
		fmt.Println("Error publishing to news:", err)
	}

	// Close the PubSub system
	if err := ps.Close(); err != nil {
		fmt.Println("Error closing pubsub:", err)
	}

	wg.Wait()
}
```

### **3️⃣ Using PubSub in a Web Server**

This example demonstrates how to use `PubSub` inside an HTTP server to push
updates.

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	ps "github.com/iamNilotpal/pubsub/pubsub"
)

var ps = pubsub.NewPubSub()

func newsHandler(w http.ResponseWriter, r *http.Request) {
	// Subscribe to the "news" topic
	ch, err := ps.Subscribe("news")
	if err != nil {
		if err == pubsub.ErrPubSubClosed {
			http.Error(w, "PubSub is closed", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "Subscription error", http.StatusInternalServerError)
		return
	}

	// Stream messages as they arrive
	for msg := range ch {
		fmt.Fprintf(w, "News update: %s\n", msg)
	}
}

func main() {
	var wg sync.WaitGroup

	// Start a simple HTTP server
	http.HandleFunc("/news", newsHandler)
	go http.ListenAndServe(":8080", nil)

	// Simulate a background process publishing news updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			ps.Publish("news", fmt.Sprintf("Breaking News %d!", i))
			time.Sleep(2 * time.Second)
		}
		ps.Close()
	}()

	wg.Wait()
}
```

👉 Run the server and visit `http://localhost:8080/news` in your browser to see
real-time updates.

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE)
file for details.

## 🌟 Show Your Support

If you like this project, give it a ⭐ on GitHub!

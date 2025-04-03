package main

import (
	"sync"
)

type PubSub struct {
	closed    bool
	mu        sync.RWMutex
	listeners map[string][]chan string
}

func NewPubSub() *PubSub {
	return &PubSub{
		mu:        sync.RWMutex{},
		listeners: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan string, 1)
	channels := ps.listeners[topic]

	ps.listeners[topic] = append(channels, ch)
	return ch
}

func (ps *PubSub) Publish(topic, msg string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}

	channels := ps.listeners[topic]
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

	for _, channels := range ps.listeners {
		for _, ch := range channels {
			close(ch)
		}
	}
}

func main() {
	pubsub := NewPubSub()
	var wg sync.WaitGroup

	devopsChan := pubsub.Subscribe("devops")
	golangChan := pubsub.Subscribe("golang")
	kubernetesChan := pubsub.Subscribe("kubernetes")

	wg.Add(3)

	go func() {
		defer wg.Done()

		for msg := range devopsChan {
			println("Message :", msg)
		}

		println("Devops channel closed")
	}()

	go func() {
		defer wg.Done()

		for msg := range golangChan {
			println("Message :", msg)
		}

		println("Golang channel closed")
	}()

	go func() {
		defer wg.Done()

		for msg := range kubernetesChan {
			println("Message :", msg)
		}

		println("Kubernetes channel closed")
	}()

	pubsub.Publish("golang", "This is a PubSub implementation using channels.")
	pubsub.Publish("devops", "DevOps Teams message.")
	pubsub.Publish("kubernetes", "Kubernetes is awesome.")

	pubsub.Close()
	pubsub.Publish("kubernetes", "k8s is so cool.")

	wg.Wait()
}

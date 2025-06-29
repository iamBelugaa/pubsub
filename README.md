# PubSub in Go

A lightweight, in-memory **Publish-Subscribe (PubSub)** system implemented in
Golang using **channels** and **goroutine**.

## 🚀 Features

- **Topic-based subscriptions**: Subscribers can listen to specific topics.
- **Channel-based communication**: Messages are sent via Go channels.
- **Thread-safe implementation**: Uses `sync.RWMutex` for concurrent access.
- **Graceful shutdown**: Closes all channels when shutting down.

## 📦 Installation

Clone the repository:

```sh
git clone https://github.com/yourusername/pubsub-go.git
cd pubsub-go
```

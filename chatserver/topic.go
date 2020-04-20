package chatserver

import (
	"sync"
)

// Represents topic that clients can subscribe to.
type Topic struct {
	name        string
	clients     map[*Client]bool
	mtx         sync.RWMutex
}

// Creates instance of a new topic.
func NewTopic(name string) *Topic {
	return &Topic{
		name,
		make(map[*Client]bool),
		sync.RWMutex{},
	}
}

// Subscribes client to a topic.
func (t *Topic) Subscribe(client *Client) {
	t.mtx.Lock()
	t.clients[client] = true
	t.mtx.Unlock()
}

// Unsubscribes client from a topic.
func (t *Topic) Unsubscribe(client *Client) {
	t.mtx.Lock()
	t.clients[client] = false
	delete(t.clients, client)
	t.mtx.Unlock()

	client.Unsubscribe()
}

// Sends a message to all clients subscribed to this topic.
func (t *Topic) SendMessage(message *Message) {
	t.mtx.RLock()

	for c, open := range t.clients {
		if open {
			c.SendMessage(message)
		}
	}

	t.mtx.RUnlock()
}

// Returns the number of subscriptions to a topic.
func (t *Topic) SubscriptionsCount() int {
	t.mtx.RLock()
	count := len(t.clients)
	t.mtx.RUnlock()

	return count
}

// Checks if this topic has subscribers.
func (t *Topic) HasSubscribers() bool {
	return t.SubscriptionsCount() > 0
}

// Unsubscribes all clients from a topic.
func (t *Topic) Destroy() {
	for client := range t.clients {
		t.Unsubscribe(client)
	}
}
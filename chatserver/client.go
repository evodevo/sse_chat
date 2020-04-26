package chatserver

import (
	"time"
)

// Represents the client connected to a chat server.
type Client struct {
	topic string
	connectedAt time.Time
	messages chan *Message
}

// Creates new client instance.
func NewClient(topic string) *Client {
	return &Client{
		topic,
		time.Now(),
		make(chan *Message),
	}
}

// Sends message to client.
func (c *Client) SendMessage(message *Message) {
	c.messages <- message
}

// Returns the topic that this client is subscribed to.
func (c *Client) GetTopic() string {
	return c.topic
}

// Returns how long the client has been connected.
func (c *Client) GetConnectedTime() time.Duration {
	return time.Since(c.connectedAt)
}

func (c *Client) Unsubscribe() {
	close(c.messages)
}

package chatserver

import (
	"fmt"
	"strings"
)

// Represents message that is sent to a client.
type Message struct {
	id string
	data string
	event string
}

// Creates text message instance.
func TextMessage(id string, data string) *Message {
	return NewMessage(id, data, "msg")
}

// Creates timeout message instance.
func TimeoutMessage(id string, data string) *Message {
	return NewMessage(id, data, "timeout")
}

// Creates new message instance.
func NewMessage(id string, data string, event string) *Message {
	return &Message{
		id,
		data,
		event,
	}
}

// Returns message as string.
func (m *Message) Serialize() string {
	var str strings.Builder

	if len(m.id) > 0 {
		str.WriteString(fmt.Sprintf("id: %s\n", m.id))
	}

	if len(m.event) > 0 {
		str.WriteString(fmt.Sprintf("event: %s\n", m.event))
	}

	if len(m.data) > 0 {
		data := strings.Replace(m.data, "\n", "\ndata: ", -1)
		str.WriteString(fmt.Sprintf("data: %s\n", data))
	}

	str.WriteString("\n")

	return str.String()
}
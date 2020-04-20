package chatserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const connectionTimeoutSec = 30

type Server struct {
	topics             map[string]*Topic
	clientConnected    chan *Client
	clientDisconnected chan *Client
	serverShutdown     chan bool
	lastMessageId      int
	logger             *log.Logger
	mtx                sync.RWMutex
}

// Creates new server instance.
func NewServer() *Server {
	s := &Server{
		make(map[string]*Topic),
		make(chan *Client),
		make(chan *Client),
		make(chan bool),
		0,
		log.New(os.Stdout, "[chatserver] ", log.LstdFlags),
		sync.RWMutex{},
	}

	go s.listen()

	return s
}

// Shuts down the server.
func (s *Server) Shutdown() {
	s.serverShutdown <- true
}

// Serves HTTP requests.
func (s *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		s.handleGetMessages(response, request)
	} else if request.Method == "POST" {
		s.handlePostMessage(response, request)
	} else {
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Handles GET messages request.
func (s *Server) handleGetMessages(response http.ResponseWriter, request *http.Request) {
	flusher, ok := response.(http.Flusher)
	if !ok {
		http.Error(response, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}

	topicName := strings.TrimPrefix(request.URL.Path, "/infocenter/")
	if topicName == "" {
		http.Error(response, "Missing topic name request parameter.", http.StatusBadRequest)
		return
	}

	client := NewClient(topicName)
	s.clientConnected <- client

	//time.AfterFunc(connectionTimeoutSec * time.Second, func() {
	//	s.onClientTimeout(client)
	//})

	client.SetTimeoutHandler(connectionTimeoutSec * time.Second, func() {
		s.onClientTimeout(client)
	})

	requestFinished := request.Context().Done()

	go func() {
		<-requestFinished
		s.clientDisconnected <- client
	}()

	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*")

	response.WriteHeader(http.StatusOK)
	flusher.Flush()

	for message := range client.messages {
		_, _ = fmt.Fprintf(response, message.Serialize())
		flusher.Flush()
	}
}

// Handles POST message request.
func (s *Server) handlePostMessage(response http.ResponseWriter, request *http.Request) {
	topicName := strings.TrimPrefix(request.URL.Path, "/infocenter/")
	if topicName == "" {
		http.Error(response, "Missing topic name request parameter.", http.StatusBadRequest)
		return
	}

	message, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(response, "Failed to read request body.", http.StatusInternalServerError)
		return
	}

	nextMessageId := s.generateMessageId()

	s.logger.Printf(
		"sending message '%s' with id %s to topic '%s'",
		message,
		strconv.Itoa(nextMessageId),
		topicName,
	)

	response.Header().Set("Access-Control-Allow-Origin", "*")

	s.sendToTopic(topicName, TextMessage(nextMessageId, string(message)))

	response.WriteHeader(http.StatusNoContent)
}

// Send message to all clients subscribed to a topic.
func (s *Server) sendToTopic(topicName string, message *Message) {
	if topic, exists := s.getTopic(topicName); exists {
		topic.SendMessage(message)

		s.logger.Printf("message sent to topic '%s'.", topicName)
	} else {
		s.logger.Printf("message not sent because topic '%s' has no subscriptions.", topicName)
	}
}

// Creates new topic on the server.
func (s *Server) createTopic(name string) *Topic {
	topic := NewTopic(name)

	s.mtx.Lock()
	s.topics[topic.name] = topic
	s.mtx.Unlock()

	s.logger.Printf("topic '%s' created.", topic.name)

	return topic
}

// Returns topic by name.
func (s *Server) getTopic(name string) (*Topic, bool) {
	s.mtx.RLock()
	topic, exists := s.topics[name]
	s.mtx.RUnlock()
	return topic, exists
}

// Destroys topic.
func (s *Server) destroyTopic(topic *Topic) {
	s.mtx.Lock()
	delete(s.topics, topic.name)
	s.mtx.Unlock()

	topic.Destroy()

	s.logger.Printf("topic '%s' destroyed.", topic.name)
}

// Destroys all topics.
func (s *Server) destroyTopics() {
	for _, topic := range s.topics {
		s.destroyTopic(topic)
	}
}

// Generates new message id.
func (s *Server) generateMessageId() int {
	s.mtx.RLock()
	s.lastMessageId++
	newId := s.lastMessageId
	s.mtx.RUnlock()

	return newId
}

// Processes server events.
func (s *Server) listen() {
	s.logger.Print("server started.")

	for {
		select {

		case c := <-s.clientConnected:
			s.onClientConnected(c)

		case c := <-s.clientDisconnected:
			s.onClientDisconnected(c)

		case <-s.serverShutdown:
			s.onServerShutdown()
			return
		}
	}
}

// Handles client connection.
func (s *Server) onClientConnected(c *Client)  {
	topic, exists := s.getTopic(c.topic)
	if !exists {
		topic = s.createTopic(c.topic)
	}

	topic.Subscribe(c)

	s.logger.Printf("client subscribed to topic '%s'.", topic.name)
}

// Handles client disconnection.
func (s *Server) onClientDisconnected(c *Client) {
	if topic, exists := s.getTopic(c.topic); exists {
		topic.Unsubscribe(c)
		s.logger.Printf("client unsubscribed from topic '%s'.", topic.name)

		if !topic.HasSubscribers() {
			s.logger.Printf("topic '%s' has no clients subscribed, destroying.", topic.name)
			s.destroyTopic(topic)
		}
	}
}

// Handles client timeout.
func (s *Server) onClientTimeout(c *Client) {
	duration := c.GetConnectedTime()
	elapsedTimeInSeconds := int(math.RoundToEven(duration.Seconds()))
	log.Printf("client connected for %d seconds, disconnecting.", elapsedTimeInSeconds)
	c.SendMessage(TimeoutMessage(
		s.generateMessageId(),
		fmt.Sprintf("%ds", elapsedTimeInSeconds),
	))
	s.clientDisconnected <- c
}

// Performs server shutdown.
func (s *Server) onServerShutdown() {
	s.destroyTopics()
	close(s.clientConnected)
	close(s.clientDisconnected)
	close(s.serverShutdown)

	s.logger.Print("server stopped.")
}
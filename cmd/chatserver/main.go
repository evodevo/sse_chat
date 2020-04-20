package main

import (
	"fmt"
	"log"
	"net/http"
	server "github.com/evodevo/sse_chat/chatserver"
)

const defaultPort = 8090

func main() {
	s := server.NewServer()
	defer s.Shutdown()

	log.Println(fmt.Sprintf("Server listening at :%d", defaultPort))
	http.Handle("/infocenter/", s)
	log.Panic(http.ListenAndServe(fmt.Sprintf(":%d", defaultPort), nil))
}
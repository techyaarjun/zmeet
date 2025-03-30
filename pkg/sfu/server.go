package sfu

import (
	"fmt"
	"net/http"
)

type Server struct {
	rooms map[string]*Room
}

func NewServer() *Server {
	return &Server{
		rooms: make(map[string]*Room),
	}
}

func (s *Server) Start() {
	fmt.Println("server started on http://localhost:9000")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.connect(w, r)
	}))
	panic(http.ListenAndServe(":9000", nil))
}

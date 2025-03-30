package sfu

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (s *Server) connect(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	name := r.URL.Query().Get("name")
	fmt.Printf("socket connecting... roomID : %v, name : %v\n", roomID, name)

	if roomID == "" || name == "" {
		fmt.Println("roomID or name is empty")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Println("socket connected.")
	defer func() {
		_ = conn.Close()
	}()

	existingRoom := s.GetRoom(roomID)
	newParticipant := NewParticipant(roomID, name, conn)
	existingRoom.Join(newParticipant)
	newParticipant.SocketConnected(true)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			return
		}

		var message Message
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println(err.Error())
			continue
		}

		//fmt.Println("message received type: ", message.Type)
		go message.process(existingRoom, newParticipant)
	}
}

func (m *Message) process(r *Room, p *Participant) {
	switch m.Type {
	case "device-ready":
		p.Init()
	case "answer":
		p.PeerAnswer(m)
	}
}

func (s *Server) GetRoom(roomID string) *Room {
	r, ok := s.rooms[roomID]
	if ok {
		return r
	}
	r = New(roomID)
	s.rooms[roomID] = r
	return r
}

package sfu

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"simple-sfu/pkg/config"
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

	if len(s.rooms) > config.MaxRooms {
		fmt.Println("max rooms reached")
		return
	}

	existingRoom := s.GetRoom(roomID)
	count := existingRoom.ParticipantCount()
	if count >= config.MaxParticipants {
		fmt.Println("max participants reached per/room")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Println("socket connected.")
	fmt.Printf("participant joined... roomID : %v, name : %v\n", roomID, name)

	newParticipant := NewParticipant(roomID, name, conn)
	existingRoom.Join(newParticipant)
	newParticipant.SocketConnected(true)

	defer func() {
		fmt.Println("socket disconnected.")
		existingRoom.Close(newParticipant)
	}()

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

		go message.process(existingRoom, newParticipant)
	}
}

func (m *Message) process(r *Room, p *Participant) {
	switch m.Type {
	case "device-ready":
		p.Init(r)
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

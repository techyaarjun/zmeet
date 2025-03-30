package sfu

import (
	"sync"
)

type Room struct {
	mu           sync.RWMutex
	id           string
	participants []*Participant
}

func New(ID string) *Room {
	return &Room{
		id:           ID,
		participants: []*Participant{},
	}
}

func (r *Room) Join(peer *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.participants = append(r.participants, peer)
}

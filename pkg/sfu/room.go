package sfu

import (
	"fmt"
	"github.com/pion/webrtc/v4"
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

func (r *Room) Close(p *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Printf("participant left... roomID : %v, name : %v\n", r.id, p.Name())

	p.mu.Lock()
	if p.peerConnected {
		err := p.peerConnection.Close()
		if err != nil {
			fmt.Println("failed to close peer connection : ", err.Error())
		}
	}
	p.mu.Unlock()

	for i, pa := range r.participants {
		if p.ID() == pa.ID() {
			r.participants = append(r.participants[:i], r.participants[i+1:]...)
			break
		}
	}

	p.mu.Lock()
	if p.socketConnected {
		err := p.socketConnection.Close()
		if err != nil {
			fmt.Println("failed to close socket connection : ", err.Error())
		}
	}
	p.mu.Unlock()
}

func (r *Room) ParticipantCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, p := range r.participants {
		p.mu.RLock()
		if p.State == CONNECTED && p.peerConnected && p.iceConnected {
			count++
		}
		p.mu.RUnlock()
	}
	return count
}

func (r *Room) GetPeerVideoOutboundTrack(myId string) (*webrtc.TrackLocalStaticRTP, *webrtc.TrackLocalStaticRTP) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, participant := range r.participants {
		participant.mu.RLock()
		if participant.id != myId {
			participant.mu.RUnlock()
			return participant.videoOutBoundTrack, participant.audioOutBoundTrack
		}
		participant.mu.RUnlock()
	}
	return nil, nil
}

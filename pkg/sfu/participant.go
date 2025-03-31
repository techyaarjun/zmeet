package sfu

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"sync"
)

type State string

var (
	CONNECTING   State = "connecting"
	READY        State = "ready"
	CONNECTED    State = "connected"
	RECONNECTING State = "reconnecting"
)

type Participant struct {
	mu                 sync.RWMutex
	id                 string
	roomID             string
	name               string
	socketConnection   *websocket.Conn
	peerConnection     *webrtc.PeerConnection
	videoOutBoundTrack *webrtc.TrackLocalStaticRTP
	audioOutBoundTrack *webrtc.TrackLocalStaticRTP
	State              State
	peerConnected      bool
	iceConnected       bool
	socketConnected    bool
}

func NewParticipant(roomID, name string, conn *websocket.Conn) *Participant {
	return &Participant{
		id:               uuid.NewString(),
		roomID:           roomID,
		name:             name,
		socketConnection: conn,
		State:            CONNECTING,
		peerConnected:    false,
		socketConnected:  false,
		iceConnected:     false,
	}
}

func (p *Participant) Init(r *Room) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peerConnection, p.videoOutBoundTrack, p.audioOutBoundTrack = NewPeerConnection()
	p.State = READY

	go MonitorTrack(p.peerConnection, p, r)
	go MonitorState(p.peerConnection, p, r)

	gatherComplete := webrtc.GatheringCompletePromise(p.peerConnection)

	offer, err := p.peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	err = p.peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	<-gatherComplete

	mess := Message{
		Type: "offer",
		Data: p.peerConnection.LocalDescription(),
	}

	_ = p.socketConnection.WriteJSON(mess)
}

func (p *Participant) SocketConnected(c bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.socketConnected = c
}

func (p *Participant) PeerConnected(c bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peerConnected = c
}

func (p *Participant) IceConnected(c bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.iceConnected = c
}

func (p *Participant) IsSocketConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.socketConnected
}

func (p *Participant) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *Participant) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

func (p *Participant) PeerAnswer(data *Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var answer webrtc.SessionDescription
	err := json.Unmarshal([]byte(data.Data.(string)), &answer)
	if err != nil {
		fmt.Println("Error unmarshalling answer")
		return
	}

	err = p.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error setting remote description")
		return
	}
}

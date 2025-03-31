package sfu

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"os"
	"simple-sfu/pkg/config"
	"strings"
	"time"
)

func NewPeerConnection() (*webrtc.PeerConnection, *webrtc.TrackLocalStaticRTP, *webrtc.TrackLocalStaticRTP) {

	settingEngine := configureSettingEngine()
	mediaEngine := configureMediaEngine()

	api := webrtc.NewAPI(
		webrtc.WithSettingEngine(settingEngine),
		webrtc.WithMediaEngine(mediaEngine),
	)

	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err.Error())
	}

	videoTack, audioTrack := addTracks(pc)

	return pc, videoTack, audioTrack
}

func MonitorTrack(pc *webrtc.PeerConnection, p *Participant, r *Room) {

	var videoOutBoundTrack *webrtc.TrackLocalStaticRTP
	var audioOutBoundTrack *webrtc.TrackLocalStaticRTP
	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					// currently the meeting is only p2p with sfu architecture
					if r.ParticipantCount() == config.MaxParticipants {
						vT, aT := r.GetPeerVideoOutboundTrack(p.ID())
						if vT == nil || aT == nil {
							fmt.Println("Video/Audio outbound track is nil")
							continue
						}
						videoOutBoundTrack = vT
						audioOutBoundTrack = aT
						ticker.Stop()
						return
					}
				}
			}
		}
	}()

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		kind := track.Kind()
		fmt.Println("Received a new track: ", kind)
		rtpBuff := make([]byte, 1500)
		rtcpBuff := make([]byte, 1500)
		PLIRequested := false

		go func() {
			for {
				_, _, err := receiver.Read(rtcpBuff)
				if err != nil {
					fmt.Println("Error reading rtcp track:", err.Error())
					return
				}
			}
		}()

		for {
			n, _, err := track.Read(rtpBuff)
			if err != nil {
				fmt.Println("Error reading rtp track:", err.Error())
				return
			}

			//pkt := &rtp.Packet{}
			//if err = pkt.Unmarshal(rtpBuff[:n]); err != nil {
			//	fmt.Println("Error parsing RTP packet: ", err.Error())
			//	continue
			//}

			if kind == webrtc.RTPCodecTypeAudio && audioOutBoundTrack != nil {
				_, err = audioOutBoundTrack.Write(rtpBuff[:n])
				if err != nil {
					fmt.Println("Error writing to audio track:", err.Error())
					continue
				}
			} else if kind == webrtc.RTPCodecTypeVideo && videoOutBoundTrack != nil {
				if !PLIRequested {
					_ = pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{
						MediaSSRC: uint32(track.SSRC()),
					}})
					PLIRequested = true
				}
				_, err = videoOutBoundTrack.Write(rtpBuff[:n])
				if err != nil {
					fmt.Println("Error writing to video track:", err.Error())
					continue
				}
			}

		}
	})
}

func MonitorState(pc *webrtc.PeerConnection, p *Participant, _ *Room) {
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		//fmt.Println("ICE Connection State changed:", state)
		if state == webrtc.ICEConnectionStateConnected {
			p.mu.Lock()
			p.iceConnected = true
			p.mu.Unlock()
		}
	})

	//pc.OnSignalingStateChange(func(state webrtc.SignalingState) {
	//	fmt.Println("Signaling State changed:", state)
	//})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		//fmt.Println("Peer Connection State changed:", state)
		if state == webrtc.PeerConnectionStateConnected {
			p.mu.Lock()
			p.State = CONNECTED
			p.peerConnected = true
			p.mu.Unlock()
		}
	})
}

func addTracks(pc *webrtc.PeerConnection) (*webrtc.TrackLocalStaticRTP, *webrtc.TrackLocalStaticRTP) {
	videoTrack, _ := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}, "video-"+uuid.NewString(), uuid.NewString())

	audioTrack, _ := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
	}, "audio-"+uuid.NewString(), uuid.NewString())

	_, _ = pc.AddTransceiverFromTrack(videoTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	})
	_, _ = pc.AddTransceiverFromTrack(audioTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	})

	return videoTrack, audioTrack

}

func configureMediaEngine() *webrtc.MediaEngine {
	mediaEngine := webrtc.MediaEngine{}
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1"},
			PayloadType:        111,
		},
	} {
		_ = mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio)
	}

	videoRTCPFeedback := []webrtc.RTCPFeedback{{"goog-remb", ""}, {"ccm", "fir"}, {"nack", ""}, {"nack", "pli"}}
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType: webrtc.MimeTypeH264, ClockRate: 90000,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 112,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeRTX, ClockRate: 90000, SDPFmtpLine: "apt=112"},
			PayloadType:        113,
		},
	} {
		_ = mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo)
	}
	return &mediaEngine
}

func configureSettingEngine() webrtc.SettingEngine {
	settingEngine := webrtc.SettingEngine{}
	ips := os.Getenv("PUBLIC_IPS")
	ipList := strings.Split(ips, ",")

	settingEngine.SetICETimeouts(4*time.Second, 12*time.Second, 2*time.Second)
	_ = settingEngine.SetEphemeralUDPPortRange(49152, 65535)
	settingEngine.SetLite(true)
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})
	settingEngine.SetNAT1To1IPs(ipList, webrtc.ICECandidateTypeHost)
	settingEngine.DisableSRTPReplayProtection(true)
	settingEngine.DisableSRTCPReplayProtection(true)
	return settingEngine
}

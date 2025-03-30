package sfu

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
	"time"
)

func NewPeerConnection() *webrtc.PeerConnection {

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

	addTracks(pc)
	go monitorState(pc)

	return pc
}

func monitorState(pc *webrtc.PeerConnection) {
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Println("ICE Connection State changed:", state)
	})

	pc.OnSignalingStateChange(func(state webrtc.SignalingState) {
		fmt.Println("Signaling State changed:", state)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Println("Peer Connection State changed:", state)
	})

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("Received a new track:", track.Kind())
	})
}

func addTracks(pc *webrtc.PeerConnection) {
	videoTrack, _ := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}, "video-"+uuid.NewString(), uuid.NewString())

	audioTrack, _ := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
	}, "audio-"+uuid.NewString(), uuid.NewString())

	_, _ = pc.AddTransceiverFromTrack(videoTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	_, _ = pc.AddTransceiverFromTrack(audioTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
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
	settingEngine.SetICETimeouts(4*time.Second, 12*time.Second, 2*time.Second)
	_ = settingEngine.SetEphemeralUDPPortRange(49152, 65535)
	settingEngine.SetLite(true)
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})
	settingEngine.SetNAT1To1IPs([]string{"192.168.1.7"}, webrtc.ICECandidateTypeHost)
	settingEngine.DisableSRTPReplayProtection(true)
	settingEngine.DisableSRTCPReplayProtection(true)
	return settingEngine
}

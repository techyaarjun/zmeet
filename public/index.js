let socket;
let pc;
let roomID;
let name;
let localStream;
let remoteVideoTrack;
let remoteAudioTrack;
const serverUrl = "ws://localhost:9000/ws";
let isMicMuted = false;
let isVideoOff = false;

async function EndButton() {
    if (confirm('Are you sure you want to end the call?')) {
        window.location.reload()
    }
}

async function MicMuteUnmute() {
    const micBtn = document.getElementById('micBtn');
    isMicMuted = !isMicMuted;
    micBtn.classList.toggle('muted', isMicMuted);

    if (isMicMuted) {
        micBtn.innerHTML = '<i class="fa-solid fa-microphone-slash"></i>';
    } else {
        micBtn.innerHTML = '<i class="fa-solid fa-microphone"></i>';
    }

    localStream.getAudioTracks().forEach(t => t.enabled = !isMicMuted)
}

async function VideoMuteUnmute() {
    const videoBtn = document.getElementById('videoBtn');
    isVideoOff = !isVideoOff;
    videoBtn.classList.toggle('muted', isVideoOff);

    if (isVideoOff) {
        videoBtn.innerHTML = '<i class="fa-solid fa-video-slash"></i>';
    } else {
        videoBtn.innerHTML = '<i class="fa-solid fa-video"></i>';
    }

    localStream.getVideoTracks().forEach(t => t.enabled = !isVideoOff)
}

async function requestMediaPermissions() {
    try {
        if (localStream) return localStream;
        localStream = await navigator.mediaDevices.getUserMedia({audio: true, video: true});
        return localStream;
    } catch (error) {
        alert("Audio and Video permissions are required!");
        return null;
    }
}

function connectWebSocket(roomID, name) {
    let url = `${serverUrl}?roomID=${roomID}&name=${name}`;
    socket = new WebSocket(url);

    socket.addEventListener("open", async () => {
        console.log("Connected to WebSocket server");
        await createPeerConnection();
        sendDeviceReady();
        document.getElementById("meeting-container").style.display = "flex";
        document.getElementById("login-container").style.display = "none";
        document.getElementById("room-id").innerHTML = "Room: " + roomID
        document.getElementById("local-name").innerHTML = name
        document.getElementById("localVideo").srcObject = localStream
    });

    socket.addEventListener("message", (event) => {
        const data = JSON.parse(event.data);

        switch (data.type) {
            case "offer": {
                handleOffer(data)
                break
            }
            default: {

            }
        }
    });

    socket.addEventListener("close", () => {
        console.log("Disconnected from WebSocket server");
        document.getElementById("meeting-container").style.display = "none";
        document.getElementById("login-container").style.display = "flex";
    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket error:", error);
        socket.close();
        document.getElementById("meeting-container").style.display = "none";
        document.getElementById("login-container").style.display = "flex";
    });
}

function sendDeviceReady() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(
            JSON.stringify({
                type: "device-ready",
                data: {
                    userAgent: navigator.userAgent,
                    language: navigator.language,
                },
            })
        );
        console.log("Device ready sent");
    }
}

async function createPeerConnection() {
    pc = await new RTCPeerConnection({});

    pc.onconnectionstatechange = () => {
        console.log("Peer Connection State:", pc.connectionState);
    };

    pc.oniceconnectionstatechange = () => {
        console.log("ICE Connection State:", pc.iceConnectionState);
    };

    pc.ontrack = (event) => {
        console.log("track received kind : ", event.track.kind)
        if (event.track.kind === "video") {
            const videoElement = document.getElementById("remoteVideo")
            if (videoElement) {
                videoElement.srcObject = event.streams[0]
                videoElement.play().catch(err => console.error(err))
            }
        } else {
            // handle
        }
    }

    const stream = await requestMediaPermissions();
    if (!stream) return;

    stream.getTracks().forEach((track) => {
        pc.addTrack(track, stream)
    })

    console.log("Peer connection created and tracks added");
}

async function handleOffer(data) {
    if (!pc) await createPeerConnection();

    const offer = data.data;
    console.log("offer : ", offer)
    await pc.setRemoteDescription(offer)
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    console.log("answer : ", answer)

    socket.send(JSON.stringify({type: "answer", data: JSON.stringify(answer)}));
}

// Generate random room ID and name
document.getElementById("roomID").value = generateRandomString(8);
document.getElementById("name").value = generateRandomName();

document.getElementById("joinForm").addEventListener("submit", async function (event) {
    event.preventDefault();
    roomID = document.getElementById("roomID").value;
    name = document.getElementById("name").value;


    if (roomID.length < 6) {
        alert("Room ID must be at least 6 characters long");
        return;
    }

    connectWebSocket(roomID, name);
});


// Generate random alphanumeric string
function generateRandomString(length) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

// Generate random name
function generateRandomName() {
    const names = [
        "Alex", "Blake", "Casey", "Dana", "Eli",
        "Fran", "Gray", "Harper", "Indigo", "Jordan",
        "Kelly", "Lee", "Morgan", "Nico", "Oakley",
        "Parker", "Quinn", "Riley", "Sky", "Taylor"
    ];
    return names[Math.floor(Math.random() * names.length)];
}

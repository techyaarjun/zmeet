let socket;
let pc;
const serverUrl = "ws://localhost:9000/ws";

async function requestMediaPermissions() {
    try {
        const stream = await navigator.mediaDevices.getUserMedia({audio: true, video: true});
        return stream;
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
    });

    socket.addEventListener("message", (event) => {
        const data = JSON.parse(event.data);
        console.log("Message from server:", data);

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
    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket error:", error);
        socket.close();
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

    const stream = await requestMediaPermissions();
    if (!stream) return;

    // Add transceivers
    const audioTransceiver = await pc.addTransceiver("audio", {direction: "sendonly"});
    const videoTransceiver = await pc.addTransceiver("video", {direction: "sendonly"});

    // Replace transceiver tracks with actual media tracks
    const audioTrack = stream.getAudioTracks()[0];
    const videoTrack = stream.getVideoTracks()[0];

    if (audioTrack) await audioTransceiver.sender.replaceTrack(audioTrack);
    if (videoTrack) await videoTransceiver.sender.replaceTrack(videoTrack);

    console.log("Peer connection created and tracks added");
}

async function handleOffer(data) {
    if (!pc) await createPeerConnection();

    const offer = data.data;
    console.log(offer)
    await pc.setRemoteDescription(offer)
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);

    socket.send(JSON.stringify({type: "answer", data: JSON.stringify(answer)}));
    console.log("Answer sent");
}

// Generate random room ID and name
document.getElementById("roomID").value = generateRandomString(8);
document.getElementById("name").value = generateRandomName();

document.getElementById("joinForm").addEventListener("submit", async function (event) {
    event.preventDefault();
    const roomID = document.getElementById("roomID").value;
    const name = document.getElementById("name").value;

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

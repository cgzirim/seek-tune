import React, { useEffect, useState } from "react";
import Peer from "simple-peer";
import io from "socket.io-client";
import Form from "./components/Form";
import Listen from "./components/Listen";
import CarouselSliders from "./components/CarouselSliders";
import AnimatedNumber from "react-animated-numbers";
import { FaMasksTheater, FaMicrophoneLines } from "react-icons/fa6";
import { LiaLaptopSolid } from "react-icons/lia";
import { ToastContainer, toast, Slide } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

// const socket = io.connect('http://localhost:5000/');
var socket = io("http://localhost:5000/");

function App() {
  const [offer, setOffer] = useState();
  const [stream, setStream] = useState();
  const [matches, setMatches] = useState([]);
  const [totalSongs, setTotalSongs] = useState(10);
  const [isListening, setisListening] = useState(false);
  const [audioInput, setAudioInput] = useState("device");
  const [peerConnection, setPeerConnection] = useState();
  const [serverEngaged, setServerEngaged] = useState(false);

  function record() {
    const mediaDevice =
      audioInput == "device"
        ? navigator.mediaDevices.getDisplayMedia.bind(navigator.mediaDevices)
        : navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);

    mediaDevice({ audio: true })
      .then(function (stream) {
        const audioTracks = stream.getAudioTracks();
        const audioStream = new MediaStream(audioTracks);

        for (const track of stream.getVideoTracks()) {
          track.stop();
        }

        const mediaRecorder = new MediaRecorder(audioStream);
        const chunks = [];

        mediaRecorder.ondataavailable = function (e) {
          chunks.push(e.data);
          console.log("DataType: ", e.data);
          // if (e.data.type.startsWith("audio")) {
          //   chunks.push(e.data);
          // }
        };

        mediaRecorder.addEventListener("stop", () => {
          const blob = new Blob(chunks, { type: "audio/mpeg" }); // Assuming MP3 format

          // Convert Blob to byte array
          const reader = new FileReader();
          reader.readAsArrayBuffer(blob);

          reader.onloadend = () => {
            if (reader.result) {
              var binary = "";
              var bytes = new Uint8Array(reader.result);
              var len = bytes.byteLength;
              for (var i = 0; i < len; i++) {
                binary += String.fromCharCode(bytes[i]);
              }

              // Convert byte array to base64
              const base64data = btoa(binary);

              console.log("Base64:", base64data);
              // Send the base64 data to the backend (assuming socket.emit exists)
              socket.emit("blob", base64data);
            } else {
              console.error("Error reading Blob as array buffer");
            }
          };
        });
        // Start recording
        mediaRecorder.start();

        // Stop recording after 5 seconds
        setTimeout(function () {
          mediaRecorder.stop();
        }, 15000);
      })
      .catch(function (err) {
        console.error("Error accessing media devices.", err);
      });
  }

  function cleanUp() {
    if (stream != null) {
      console.log("Cleaning tracks");
      stream.getTracks().forEach((track) => track.stop());
    }
    setStream(null);
    setisListening(false);
    console.log("Cleanup complete.");
  }

  function createPeerConnection() {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      stream: null,
    });

    // Handle peer events:
    peer.on("signal", (offerData) => {
      console.log("Setting Offer!");
      setOffer(JSON.stringify(offerData));
      setPeerConnection(peer);
    });

    peer.on("close", () => {
      cleanUp();
      setServerEngaged(false);
      console.log("CONNECTION CLOSED");
    });

    peer.on("error", (err) => {
      console.error("An error occurred:", err);
    });
  }

  useEffect(() => {
    if (offer) {
      console.log("Offer updated:", offer);
      let offerEncoded = btoa(offer);
      socket.emit("engage", offerEncoded);

      socket.on("serverEngaged", (answer) => {
        console.log("ServerSDP: ", answer);
        let decodedAnswer = atob(answer);
        if (!serverEngaged && !stream && !peerConnection.destroyed) {
          peerConnection.signal(decodedAnswer);
        }
        console.log("Engaged Server");
        setServerEngaged(true);
      });
    }
  }, [offer]);

  useEffect(() => {
    socket.on("connect", () => {
      createPeerConnection();
    });

    socket.on("matches", (matches) => {
      matches = JSON.parse(matches);
      if (matches) {
        setMatches(matches);
        console.log("Matches: ", matches);
      } else {
        console.log("No Matches");
      }

      cleanUp();
    });

    socket.on("downloadStatus", (msg) => {
      console.log("downloadStatus: ", msg);
      msg = JSON.parse(msg);
      const msgTypes = ["info", "success", "error"];
      if (msg.type !== undefined && msgTypes.includes(msg.type)) {
        toast[msg.type](() => <div>{msg.message}</div>);
      } else {
        toast(msg.message);
      }
    });

    socket.on("totalSongs", (songsCount) => {
      console.log("Total songs in DB: ", songsCount);
      setTotalSongs(songsCount);
    });
  }, []);

  useEffect(() => {
    const emitTotalSongs = () => {
      socket.emit("totalSongs", "");
    };

    const intervalId = setInterval(emitTotalSongs, 8000);

    return () => clearInterval(intervalId);
  }, []);

  function stopListening() {
    console.log("Pause Clicked");
    cleanUp();
    peerConnection.destroy();

    setTimeout(() => {
      createPeerConnection();
    }, 3);
  }

  function startListening() {
    const mediaDevice =
      audioInput === "device"
        ? navigator.mediaDevices.getDisplayMedia.bind(navigator.mediaDevices)
        : navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);

    mediaDevice({ audio: true })
      .then((stream) => {
        console.log("isListening: ", isListening);
        peerConnection.addStream(stream);
        setisListening(true);

        setStream(stream);
        stream.getVideoTracks()[0].onended = stopListening;
        stream.getAudioTracks()[0].onended = stopListening;
      })
      .catch((error) => {
        console.error("Error accessing user media:", error);
      });
  }

  const handleLaptopIconClick = () => {
    console.log("Laptop icon clicked");
    setAudioInput("device");
  };

  const handleMicrophoneIconClick = () => {
    console.log("Microphone icon clicked");
    setAudioInput("mic");
  };

  return (
    <div className="App">
      <h1>New App</h1>
      <h4 style={{ display: "flex", justifyContent: "flex-end" }}>
        <AnimatedNumber
          includeComma
          animateToNumber={totalSongs}
          config={{ tension: 89, friction: 40 }}
          animationType={"calm"}
        />
        &nbsp;Songs
      </h4>
      <div className="listen">
        <Listen
          stopListening={stopListening}
          disable={!serverEngaged}
          startListening={startListening}
          isListening={isListening}
        />
      </div>
      <div className="audio-input">
        <div
          onClick={handleLaptopIconClick}
          className={
            audioInput !== "device"
              ? "audio-input-device"
              : "audio-input-device active-audio-input"
          }
        >
          <LiaLaptopSolid style={{ height: 20, width: 20 }} />
        </div>
        <div
          onClick={handleMicrophoneIconClick}
          className={
            audioInput !== "mic"
              ? "audio-input-mic"
              : "audio-input-mic active-audio-input"
          }
        >
          <FaMicrophoneLines style={{ height: 20, width: 20 }} />
        </div>
      </div>
      <div className="youtube">
        <CarouselSliders matches={matches} />
      </div>
      <Form socket={socket} toast={toast} />
      <ToastContainer
        position="top-center"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="light"
        transition={Slide}
      />
    </div>
  );
}

export default App;

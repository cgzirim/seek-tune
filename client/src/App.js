import React, { useEffect, useState, useRef } from "react";
import io from "socket.io-client";
import Form from "./components/Form";
import Listen from "./components/Listen";
import CarouselSliders from "./components/CarouselSliders";
import { FaMicrophoneLines } from "react-icons/fa6";
import { LiaLaptopSolid } from "react-icons/lia";
import { ToastContainer, toast, Slide } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { MediaRecorder, register } from "extendable-media-recorder";
import { connect } from "extendable-media-recorder-wav-encoder";

import AnimatedNumber from "./components/AnimatedNumber";

const server = process.env.REACT_APP_BACKEND_URL || "http://localhost:5000";

var socket = io(server);

function App() {
  const [stream, setStream] = useState();
  const [matches, setMatches] = useState([]);
  const [totalSongs, setTotalSongs] = useState(10);
  const [isListening, setisListening] = useState(false);
  const [audioInput, setAudioInput] = useState("device"); // or "mic"
  const [isPhone, setIsPhone] = useState(window.innerWidth <= 550);
  const [registeredMediaEncoder, setRegisteredMediaEncoder] = useState(false);

  const streamRef = useRef(stream);
  let sendRecordingRef = useRef(true);

  useEffect(() => {
    streamRef.current = stream;
  }, [stream]);

  useEffect(() => {
    if (isPhone) {
      setAudioInput("mic");
    }

    socket.on("connect", () => {
      socket.emit("totalSongs", "");
    });

    socket.on("matches", (matches) => {
      matches = JSON.parse(matches);
      if (matches) {
        setMatches(matches.slice(0, 5));
        console.log("Matches: ", matches);
      } else {
        toast("No song found.");
      }

      cleanUp();
    });

    socket.on("downloadStatus", (msg) => {
      msg = JSON.parse(msg);
      const msgTypes = ["info", "success", "error"];
      if (msg.type !== undefined && msgTypes.includes(msg.type)) {
        toast[msg.type](() => <div>{msg.message}</div>);
      } else {
        toast(msg.message);
      }
    });

    socket.on("totalSongs", (songsCount) => {
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

  async function record() {
    try {
      const mediaDevice =
        audioInput === "device"
          ? navigator.mediaDevices.getDisplayMedia.bind(navigator.mediaDevices)
          : navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);

      if (!registeredMediaEncoder) {
        await register(await connect());
        setRegisteredMediaEncoder(true);
      }

      const constraints = {
        audio: {
          autoGainControl: false,
          channelCount: 1,
          echoCancellation: false,
          noiseSuppression: false,
          sampleSize: 16,
        },
      };

      const stream = await mediaDevice(constraints);
      const audioTracks = stream.getAudioTracks();
      const audioStream = new MediaStream(audioTracks);

      setStream(audioStream);

      audioTracks[0].onended = stopListening;

      // Stop video tracks
      for (const track of stream.getVideoTracks()) {
        track.stop();
      }

      /** Attempt to change sampleRate
      const audioContext = new AudioContext({
        sampleRate: 44100,
      });
      const mediaStreamAudioSourceNode = new MediaStreamAudioSourceNode(
        audioContext,
        { mediaStream: audioStream }
      );
      const mediaStreamAudioDestinationNode =
        new MediaStreamAudioDestinationNode(audioContext, {
          channelCount: 1,
        });

      mediaStreamAudioSourceNode.connect(mediaStreamAudioDestinationNode);

      const mediaRecorder = new MediaRecorder(
        mediaStreamAudioDestinationNode.stream,
        { mimeType: "audio/wav" }
      );

      const settings = mediaStreamAudioDestinationNode.stream
        .getAudioTracks()[0]
        .getSettings();

      console.log("Settings: ", settings);
      */

      const mediaRecorder = new MediaRecorder(audioStream, {
        mimeType: "audio/wav",
      });

      mediaRecorder.start();
      setisListening(true);
      sendRecordingRef.current = true;

      const chunks = [];
      mediaRecorder.ondataavailable = function (e) {
        chunks.push(e.data);
      };

      // Stop recording after 20 seconds
      setTimeout(function () {
        mediaRecorder.stop();
      }, 20000);

      mediaRecorder.addEventListener("stop", () => {
        const blob = new Blob(chunks, { type: "audio/wav" });
        const reader = new FileReader();

        cleanUp();
        // downloadRecording(blob);

        reader.readAsArrayBuffer(blob);
        reader.onload = async (event) => {
          const arrayBuffer = event.target.result;

          // get record duration
          const arrayBufferCopy = arrayBuffer.slice(0);
          const audioContext = new AudioContext();
          const audioBufferDecoded = await audioContext.decodeAudioData(
            arrayBufferCopy
          );
          const recordDuration = audioBufferDecoded.duration;

          var binary = "";
          var bytes = new Uint8Array(arrayBuffer);
          var len = bytes.byteLength;
          for (var i = 0; i < len; i++) {
            binary += String.fromCharCode(bytes[i]);
          }

          // Convert byte array to base64
          const rawAudio = btoa(binary);
          const audioConfig = audioStream.getAudioTracks()[0].getSettings();

          const recordData = {
            audio: rawAudio,
            duration: recordDuration,
            channels: audioConfig.channelCount,
            sampleRate: audioConfig.sampleRate,
            sampleSize: audioConfig.sampleSize,
          };

          if (sendRecordingRef.current) {
            socket.emit("newRecording", JSON.stringify(recordData));
          }
        };
      });
    } catch (error) {
      console.error("error:", error);
      cleanUp();
    }
  }

  function downloadRecording(blob) {
    const blobUrl = URL.createObjectURL(blob);

    // Create a download link
    const downloadLink = document.createElement("a");
    downloadLink.href = blobUrl;
    downloadLink.download = "recorded_audio.wav";
    document.body.appendChild(downloadLink);
    downloadLink.click();
  }

  function cleanUp() {
    const currentStream = streamRef.current;
    if (currentStream) {
      currentStream.getTracks().forEach((track) => track.stop());
    }

    setStream(null);
    setisListening(false);
  }

  function stopListening() {
    cleanUp();
    sendRecordingRef.current = false;
  }

  function handleLaptopIconClick() {
    setAudioInput("device");
  }

  function handleMicrophoneIconClick() {
    setAudioInput("mic");
  }

  return (
    <div className="App">
      <div className="TopHeader">
        <h2 style={{ color: "#374151" }}>SeekTune</h2>
        <h4 style={{ display: "flex", justifyContent: "flex-end" }}>
          <AnimatedNumber includeComma={true} animateToNumber={totalSongs} />
          &nbsp;Songs
        </h4>
      </div>
      <div className="listen">
        <Listen
          stopListening={stopListening}
          disable={false}
          startListening={record}
          isListening={isListening}
        />
      </div>
      {!isPhone && (
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
      )}
      <div className="youtube">
        <CarouselSliders matches={matches} />
      </div>
      <Form socket={socket} toast={toast} />
      <ToastContainer
        position="top-center"
        autoClose={5000}
        hideProgressBar={true}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        pauseOnHover
        theme="light"
        transition={Slide}
      />
    </div>
  );
}

export default App;

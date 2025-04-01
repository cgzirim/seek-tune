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
import { FFmpeg } from '@ffmpeg/ffmpeg';
import { fetchFile } from '@ffmpeg/util';


import AnimatedNumber from "./components/AnimatedNumber";

const server = process.env.REACT_APP_BACKEND_URL || "http://localhost:5000";
// https://seek-tune-rq4gn.ondigitalocean.app/

var socket = io(server);

function App() {
  let ffmpegLoaded = false;
  const ffmpeg = new FFmpeg();
  const uploadRecording = true
  const isPhone = window.innerWidth <= 550
  const [stream, setStream] = useState();
  const [matches, setMatches] = useState([]);
  const [totalSongs, setTotalSongs] = useState(10);
  const [isListening, setisListening] = useState(false);
  const [audioInput, setAudioInput] = useState("device"); // or "mic"
  const [genFingerprint, setGenFingerprint] = useState(null);
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

  useEffect(() => { 
    (async () => {
      try {
        const go = new window.Go();
        const result = await WebAssembly.instantiateStreaming(
          fetch("/main.wasm"), 
          go.importObject
        );
        go.run(result.instance);

        if (typeof window.generateFingerprint === "function") {
          setGenFingerprint(() => window.generateFingerprint);
        }

      } catch (error) {
        console.error("Error loading WASM:", error);
      }
    })();
  }, []);

  async function record() {
    try {
      if (!genFingerprint) {
        console.error("WASM is not loaded yet.");
        return;
      }

      if (!ffmpegLoaded) {
        await ffmpeg.load();
        ffmpegLoaded = true;
      }

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

      mediaRecorder.addEventListener("stop", async () => {
        const blob = new Blob(chunks, { type: "audio/wav" });

        cleanUp();

        const inputFile = 'input.wav';
        const outputFile = 'output_mono.wav';

        // Convert audio to mono with a sample rate of 44100 Hz
        await ffmpeg.writeFile(inputFile, await fetchFile(blob))
        const exitCode = await ffmpeg.exec([
          '-i', inputFile,
          '-c', 'pcm_s16le',
          '-ar', '44100',
          '-ac', '1',
          '-f', 'wav',
          outputFile
        ]);
        if (exitCode !== 0) {
          throw new Error(`FFmpeg exec failed with exit code: ${exitCode}`);
        }

        const monoData = await ffmpeg.readFile(outputFile);
        const monoBlob = new Blob([monoData.buffer], { type: 'audio/wav' });

        const reader = new FileReader();
        reader.readAsArrayBuffer(monoBlob);
        reader.onload = async (event) => {
          const arrayBuffer = event.target.result;
          const audioContext = new AudioContext();
          const arrayBufferCopy = arrayBuffer.slice(0);
          const audioBufferDecoded = await audioContext.decodeAudioData(arrayBufferCopy);
          
          const audioData = audioBufferDecoded.getChannelData(0);
          const audioArray = Array.from(audioData);

          const result = genFingerprint(audioArray, audioBufferDecoded.sampleRate);
          if (result.error !== 0) {
            toast["error"](() => <div>An error occured</div>)
            console.log("An error occured: ", result)
            return
          }

          const fingerprintMap = result.data.reduce((dict, item) => {
            dict[item.address] = item.anchorTime;
            return dict;
          }, {});

          if (sendRecordingRef.current) {
            socket.emit("newFingerprint", JSON.stringify({ fingerprint: fingerprintMap }));
          }

          if (uploadRecording) {
            var bytes = new Uint8Array(arrayBuffer);
            var rawAudio = "";
            for (var i = 0; i < bytes.byteLength; i++) {
              rawAudio += String.fromCharCode(bytes[i]);
            }

            const dataView = new DataView(arrayBuffer);

            const recordData = {
              audio: btoa(rawAudio),
              channels: dataView.getUint16(22, true),
              sampleRate: dataView.getUint16(24, true),
              sampleSize: dataView.getUint16(34, true),
              duration: audioBufferDecoded.duration,
            };

            console.log("Record data: ", recordData);

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
        <h2 style={{ color: "#374151" }}>!Shazam</h2>
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
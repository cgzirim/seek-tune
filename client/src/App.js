import React, { useEffect, useState } from "react";
import Peer from "simple-peer";
import io, { managers } from "socket.io-client";
import Form from "./Form";

// const socket = io.connect('http://localhost:5000/');
var socket = io("http://localhost:5000/");

function App() {
  const [offer, setOffer] = useState();
  const [stream, setStream] = useState();
  const [matches, setMatches] = useState([]);
  const [serverEngaged, setServerEngaged] = useState(false);
  const [peerConnection, setPeerConnection] = useState();

  // Function to initiate the client peer
  function initiateClientPeer(stream = null) {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      stream: stream,
    });

    let offerHasBeenSet = false;

    peer.on("signal", (data) => {
      if (!offerHasBeenSet) {
        console.log("Setting Offer!");
        setOffer(JSON.stringify(data));
        offerHasBeenSet = true;
      }
    });

    peer.on("close", () => {
      console.log("CONNECTION CLOSED");
    });

    peer.on("error", (err) => {
      console.error("An error occurred:", err);
    });

    setPeerConnection(peer);
  }

  useEffect(() => {
    console.log("Offer updated:", offer);
    let renegotiated = false;

    if (offer && stream && !renegotiated) {
      let offerEncoded = btoa(offer);
      socket.emit("engage", offerEncoded);

      socket.on("serverEngaged", (answer) => {
        console.log("ServerSDP: ", answer);
        let decodedAnswer = atob(answer);
        peerConnection.signal(decodedAnswer);
        console.log("Engaging Server");
        setServerEngaged(true);
        renegotiated = true;
      });
    }
  }, [offer]);

  useEffect(() => {
    initiateClientPeer();
  }, []);

  // socket.on("connect", () => {
  //   initiateClientPeer();
  // });

  socket.on("matches", (matches) => {
    matches = JSON.parse(matches);
    setMatches(matches);
    console.log("Matches: ", matches);
  });

  socket.on("downloadStatus", (msg) => {
    console.log("downloadStatus: ", msg);
  });

  socket.on("albumStat", (msg) => {
    console.log("Album stat: ", msg);
  });

  socket.on("playlistStat", (msg) => {
    console.log("Playlist stat: ", msg);
  });

  const streamAudio = () => {
    navigator.mediaDevices
      .getDisplayMedia({ audio: true })
      .then((stream) => {
        peerConnection.addStream(stream);

        // Renegotiate
        let initOfferEncoded = btoa(offer);
        socket.emit("initOffer", initOfferEncoded);

        socket.on("initAnswer", (answer) => {
          let decodedAnswer = atob(answer);
          peerConnection.signal(decodedAnswer);
          console.log("Renogotiated");
        });

        // End of Renegotiation

        peerConnection.on("signal", (data) => {
          setOffer(JSON.stringify(data));
          console.log("Offer should be reset");
        });
        setStream(stream); // Set the audio stream to state
      })
      .catch((error) => {
        console.error("Error accessing user media:", error);
        // Handle error
      });

    if (!offer || !peerConnection) {
      // If offer is not set, create a new one
      console.log("NO OFFER. CREATING OFFER");
      initiateClientPeer(stream);
    }
  };

  const disengageServer = () => {
    peerConnection.destroy();
  };

  return (
    <div className="App">
      <h1>New App</h1>
      <div>
        {serverEngaged ? (
          <button disabled={true}>Listening</button>
        ) : (
          <button onClick={() => streamAudio()}>Listen</button>
        )}
        {serverEngaged && (
          <button onClick={() => disengageServer()}>Stop Listening</button>
        )}
      </div>
      <Form socket={socket} />
      <div>
        {matches.map((match, index) => {
          const [h, m, s] = match.timestamp.split(":");
          const timestamp =
            parseInt(h, 10) * 120 + parseInt(m, 10) * 60 + parseInt(s, 10);

          return (
            <iframe
              key={index}
              width="460"
              height="284"
              src={`https://www.youtube.com/embed/${match.youtubeid}?start=${timestamp}`}
              title={match.songname}
              frameBorder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
              allowFullScreen
            ></iframe>
          );
        })}
      </div>
    </div>
  );
}

export default App;

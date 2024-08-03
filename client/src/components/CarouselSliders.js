import React, { useEffect, useRef, useState } from "react";
import YouTube from "react-youtube";
import styles from "./styles/CarouselSliders.module.css";

const CarouselSliders = (props) => {
  const [activeVideoID, setActiveVideoID] = useState(null);
  const players = useRef({});

  useEffect(() => {
    if (props.matches.length > 0) {
      // Filter out matches with empty YouTubeID
      const validMatches = props.matches.filter((match) => match.YouTubeID);
  
      if (validMatches.length > 0) {
        const firstVideoID = validMatches[0].YouTubeID;
        document
          .getElementById(`slide-${firstVideoID}`)
          .scrollIntoView({ behavior: "smooth" });
        setActiveVideoID(firstVideoID);
      }
    }
  }, [props.matches]);

  const onReady = (event, videoId) => {
    players.current[videoId] = event.target;
  };

  const onPlay = (event) => {
    const videoId = event.target.getVideoData().video_id;
    setActiveVideoID(videoId);

    // Pause other videos
    Object.values(players.current).forEach((player) => {
      const otherVideoId = player.getVideoData().video_id;
      if (
        otherVideoId !== videoId &&
        player.getPlayerState() === 1 /* Playing */
      ) {
        player.pauseVideo();
      }
    });
  };

  return (
    <>
      <div className={styles.CarouselSliders}>
        {!props.matches.length ? null : (
          <div className={styles.Slider}>
            {props.matches
            .filter((match) => match.YouTubeID) // Filter out matches with empty YouTubeID
            .map((match, index) => {
              const start = (parseInt(match.Timestamp) / 1000) | 0;

              return (
                <div
                  key={index}
                  id={`slide-${match.YouTubeID}`}
                  className={styles.SlideItem}
                >
                  <YouTube
                    videoId={match.YouTubeID}
                    opts={{
                      playerVars: { start: start, rel: 0 },
                    }}
                    iframeClassName={styles.Iframe}
                    onReady={(event) => onReady(event, match.YouTubeID)}
                    onPlay={onPlay}
                  />
                </div>
              );
            })}
          </div>
        )}

        <div className={styles.Circles}>
          {props.matches
          .filter((match) => match.YouTubeID)
          .map((match, _) => {
            return (
              <a
                key={match.YouTubeID}
                className={
                  match.YouTubeID !== activeVideoID
                    ? styles.Link
                    : `${styles.Link} ${styles.ActiveLink}`
                }
                href={`#slide-${match.YouTubeID}`}
                onClick={(e) => {
                  e.preventDefault();
                  document
                    .getElementById(`slide-${match.YouTubeID}`)
                    .scrollIntoView({ behavior: "smooth" });
                  setActiveVideoID(match.YouTubeID);
                }}
              ></a>
            );
          })}
        </div>
      </div>
    </>
  );
};

export default CarouselSliders;

import React, { useRef } from "react";
import YouTube from "react-youtube";
import styles from "./styles/CarouselSliders.module.css";

const CarouselSliders = (props) => {
  const [activeIdx, setActiveIdx] = React.useState(0);
  const players = useRef({});

  const opts = {
    // width: "420",
    // height: "210",
  };

  const onReady = (event, videoId) => {
    players.current[videoId] = event.target;
  };

  const onPlay = (event) => {
    const videoId = event.target.getVideoData().video_id;
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
            {props.matches.map((match, index) => {
              const [h, m, s] = match.timestamp.split(":");
              const timestamp =
                parseInt(h, 10) * 360 + parseInt(m, 10) * 60 + parseInt(s, 10);

              return (
                <div
                  key={index}
                  id={`slide-${index}`}
                  className={styles.SlideItem}
                >
                  <YouTube
                    videoId={match.youtubeid}
                    opts={{
                      ...opts,
                      playerVars: { start: timestamp, rel: 0 },
                    }}
                    iframeClassName={styles.Iframe}
                    onReady={(event) => onReady(event, match.youtubeid)}
                    onPlay={onPlay}
                  />
                </div>
              );
            })}
          </div>
        )}

        <div className={styles.Circles}>
          {props.matches.map((_, index) => {
            return (
              <a
                key={index}
                className={
                  index !== activeIdx
                    ? styles.Link
                    : `${styles.Link} ${styles.ActiveLink}`
                }
                href={`#slide-${index}`}
                onClick={() => {
                  setActiveIdx(index);
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

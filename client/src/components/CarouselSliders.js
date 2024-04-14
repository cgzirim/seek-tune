import React from "react";
import styles from "./styles/CarouselSliders.module.css";

const CarouselSliders = (props) => {
  const [activeIdx, setActiveIdx] = React.useState(0);

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
                <>
                  <div
                    key={index}
                    id={`slide-${index}`}
                    className={styles.SlideItem}
                  >
                    <iframe
                      className="iframe-youtube"
                      src={`https://www.youtube.com/embed/${match.youtubeid}?start=${timestamp}`}
                      title={match.songname}
                      frameBorder="0"
                      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                      allowFullScreen
                    ></iframe>
                  </div>
                </>
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

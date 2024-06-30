import React from "react";
import { motion } from "framer-motion";

const NUMBERS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];

const usePrevious = (value) => {
  const ref = React.useRef();
  React.useEffect(() => {
    ref.current = value;
  });

  if (typeof ref.current === "undefined") {
    return 0;
  }

  return ref.current;
};

const AnimatedNumber = ({
  animateToNumber,
  fontStyle,
  transition = { duration: 0.8 },
  includeComma,
  delay,
  onFinish,
  onStart,
}) => {
  const prevNumber = usePrevious(animateToNumber);
  const animateToNumberString = String(Math.abs(animateToNumber));
  const prevNumberString = String(Math.abs(prevNumber));
  const animateToNumbersArr = Array.from(animateToNumberString, Number);
  const prevNumbersArr = Array.from(prevNumberString, Number);
  const [onStarted, setOnStarted] = React.useState(false);
  const [onFinished, setOnFinished] = React.useState(false);

  React.useEffect(() => {
    if (onStarted && onStart) {
      const delayTime = typeof delay === "undefined" ? 0 : delay;

      setTimeout(() => {
        onStart();
      }, delayTime);
    }
  }, [onStarted]);

  React.useEffect(() => {
    if (onFinished && onFinish) {
      onFinish();
    }
  }, [onFinished]);

  if (includeComma) {
    const reducedArray = new Array(
      Math.ceil(animateToNumberString.length / 3)
    ).fill(0);

    const startReducedArray = new Array(
      Math.ceil(prevNumberString.length / 3)
    ).fill(0);

    reducedArray.map((__, index) => {
      if (index === 0) {
        return;
      }

      animateToNumbersArr.splice(
        animateToNumberString.length - index * 3,
        0,
        ","
      );
    });

    startReducedArray.map((__, index) => {
      if (index === 0) {
        return;
      }

      prevNumbersArr.splice(prevNumberString.length - index * 3, 0, ",");
    });
  }

  const [numberHeight, setNumberHeight] = React.useState(0);

  const numberDivRef = React.useRef(null);

  React.useEffect(() => {
    setNumberHeight(numberDivRef.current.clientHeight);
  }, [animateToNumber]);

  return (
    <>
      {numberHeight !== 0 && (
        <div style={{ display: "flex", flexDirection: "row" }}>
          {animateToNumber < 0 && <div style={fontStyle}>-</div>}
          {animateToNumbersArr.map((n, index) => {
            if (typeof n === "string") {
              return (
                <div key={index} style={{ ...fontStyle }}>
                  {n}
                </div>
              );
            }

            return (
              <div
                key={index}
                style={{
                  height: numberHeight,
                  overflow: "hidden",
                }}
              >
                <motion.div
                  initial={{ y: 0 }}
                  animate={{
                    y: -1 * (numberHeight * animateToNumbersArr[index]),
                  }}
                  transition={transition}
                  onAnimationStart={() => setOnStarted(true)}
                  onAnimationComplete={() => setOnFinished(true)}
                >
                  {NUMBERS.map((number, i) => (
                    <div key={i} style={fontStyle}>
                      {number}
                    </div>
                  ))}
                </motion.div>
              </div>
            );
          })}
        </div>
      )}

      <div
        ref={numberDivRef}
        style={{ position: "absolute", top: -9999, ...fontStyle }}
      >
        {0}
      </div>
    </>
  );
};

export default AnimatedNumber;

import React, { useState } from "react";
import AnimatedNumber from "./AnimatedNumber";
import { Bounce } from "react-toastify";

const TestComponent = (props) => {
  const [number, setNumber] = useState(18);

  const increment = () => setNumber((prev) => prev + 1);
  const decrement = () => setNumber((prev) => prev - 1);

  return (
    <div>
      <h1>Animated Number</h1>
      <AnimatedNumber
        animateToNumber={number}
        fontStyle={{ fontSize: "2rem", color: "black" }}
        // transitions={(index) => ({
        //   duration: 10,
        //   delay: index * 10,
        // })}
        includeComma={true}
      />
      <div>
        <button onClick={increment}>Increment</button>
        <button onClick={decrement}>Decrement</button>
      </div>
    </div>
  );
};

export default TestComponent;

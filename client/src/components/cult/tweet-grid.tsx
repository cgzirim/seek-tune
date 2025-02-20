"use client"

import * as React from "react"
import { Tweet } from "react-tweet"

import { GradientHeading } from "./gradient-heading"

const tweets = [
  "1742983975340327184",
  "1743049700583116812",
  "1754067409366073443",
  "1753968111059861648",
  "1754174981897118136",
  "1743632296802988387",
  "1754110885168021921",
  "1760248682828419497",
  "1760230134601122153",
  "1760184980356088267",
  "1742983975340327184",
  "1742983975340327184",
  "1743049700583116812",
]

export function TweetGrid({}) {
  return (
    <div className="pb-12 ">
      <div className="flex w-full justify-center pb-12">
        <GradientHeading size="xl" weight="black">
          Join the club
        </GradientHeading>
      </div>
      <div className="flex items-center justify-center w-full">
        <div className="columns-1 sm:columns-2 md:columns-4  max-w-4xl md:max-w-6xl px-2">
          {tweets.map((tweetId, i) => (
            <div key={`${tweetId}-${i}`} className="mb-4">
              <Tweet id={tweetId} />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

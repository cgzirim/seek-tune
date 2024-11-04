"use client"

import * as React from "react"
import Image, { StaticImageData } from "next/image"
import lowes1 from "@/images/projects/Clay-1.png"
import merc3 from "@/images/projects/Clay-2.png"
import lowes2 from "@/images/projects/Clay-3.png"
import allyMilestone2 from "@/images/projects/Clay-4.png"
import allyMilestone from "@/images/projects/Clay-5.png"
import merc4 from "@/images/projects/Clay-6.png"
import allyArena1 from "@/images/projects/img-1.png"
import allyArena2 from "@/images/projects/img-2.png"
import allyArena3 from "@/images/projects/img-3.png"
import {
  motion,
  MotionValue,
  useAnimationFrame,
  useMotionTemplate,
  useMotionValue,
  useTransform,
} from "framer-motion"

import { cn } from "@/lib/utils"

import { CameraTarget, useCamera, utils } from "./camera"

const bannerTwoImage2s = [allyMilestone, merc4, allyArena1, lowes1, allyArena3]

interface InfiniteBannerProps extends React.HTMLProps<HTMLDivElement> {
  clock: MotionValue<number>
  loopDuration?: number
  children: React.ReactNode
}

const InfiniteBanner = ({
  clock,
  loopDuration = 22000,
  children,
  ...otherProps
}: InfiniteBannerProps) => {
  const progress = useTransform(
    clock,
    (time) => (time % loopDuration) / loopDuration
  )
  const percentage = useTransform(progress, (t) => t * 100)
  const translateX = useMotionTemplate`-${percentage}%`
  return (
    <div
      {...otherProps}
      style={{
        ...otherProps.style,
      }}
      className="relative w-max overflow-hidden"
    >
      <motion.div style={{ translateX, width: "max-content" }}>
        <div>{children}</div>
        <div className="absolute h-full w-full left-full top-0">{children}</div>
      </motion.div>
    </div>
  )
}

export const useClock = ({
  defaultValue = 0,
  reverse = false,
  speed = 1,
} = {}) => {
  const clock = useMotionValue(defaultValue)
  const paused = React.useRef(false)
  useAnimationFrame((t, dt) => {
    if (paused.current) {
      return
    }
    if (reverse) {
      clock.set(clock.get() - dt * speed)
    } else {
      clock.set(clock.get() + dt * speed)
    }
  })
  return {
    value: clock,
    stop: () => {
      paused.current = true
    },
    start: () => {
      paused.current = false
    },
  }
}

// const bannerOneImages = [
//   '/images/projects/ally-arena-1.png',
//   '/images/projects/lowes-2.png',
//   '/images/projects/merc-3.png',
//   '/images/projects/ally-arena-2.png',
//   '/images/projects/ally-milestone-2.png',
// ]

const bannerOneImages = [
  // '/images/projects/ally-arena-1.png',
  allyArena1,
  lowes2,
  merc3,
  allyArena2,
  allyMilestone2,
  // '/images/projects/lowes-2.png',
  // '/images/projects/merc-3.png',
  // '/images/projects/ally-arena-2.png',
  // '/images/projects/ally-milestone-2.png',
]

const bannerTwoImages = [allyMilestone, merc4, allyArena1, lowes1, allyArena3]
// const bannerTwoImages = [
//   '/images/projects/ally-milestone.png',
//   '/images/projects/merc-4.png',
//   '/images/projects/ally-arena-1.png',
//   '/images/projects/lowes-1.png',
//   '/images/projects/ally-arena-3.png',
// ]

interface PhotoProps {
  src: StaticImageData
  onClick: (target: utils.CameraTarget | null) => void
}

const Photo = ({ src, onClick }: PhotoProps) => {
  const ref = React.useRef<utils.CameraTarget>(null)
  const [isFull, setIsFull] = React.useState(false)

  return (
    <CameraTarget ref={ref}>
      <Image
        tabIndex={0}
        src={src}
        alt={"project-image"}
        placeholder="blur"
        // width={450}
        // height={300}
        onClick={() => {
          onClick(ref.current)
          setIsFull((isFull) => !isFull)
        }}
        className={cn(
          "cursor-pointer  border-8 border-black rounded-xl",
          isFull
            ? "h-64 w-auto md:w-[450px] md:h-[300px] object-cover aspect-auto"
            : "h-64 w-auto md:w-[450px] md:h-[300px]  object-cover aspect-video"
        )}
      />
    </CameraTarget>
  )
}

const Banners = () => {
  const camera = useCamera()
  const [target, setTarget] = React.useState<utils.CameraTarget>(null)
  const clock = useClock({
    defaultValue: Date.now(),
    reverse: false,
  })
  const reverseClock = useClock({
    defaultValue: Date.now(),
    reverse: true,
  })

  // React.useEffect(() => {
  //   if (target) {
  //     camera.follow(target)
  //     camera.setZoom(1.55)
  //     camera.setRotation(0)
  //     clock.stop()
  //     reverseClock.stop()
  //   } else {
  //     camera.panTo(new utils.Vector(0, 0))
  //     camera.setZoom(1)
  //     camera.setRotation(-10)
  //     clock.start()
  //     reverseClock.start()
  //   }
  //   return () => {
  //     if (target) camera.unfollow(target)
  //   }
  // }, [camera, target, clock, reverseClock])

  camera.setRotation(-10)

  return (
    <div className="space-y-6">
      <InfiniteBanner clock={clock.value}>
        <div className="flex space-x-6 pr-6 ">
          {bannerOneImages.map((img, i) => (
            <Photo
              key={`set1-${i}`}
              src={img}
              onClick={(t) => setTarget((prev) => (prev !== t ? t : null))}
            />
          ))}
        </div>
      </InfiniteBanner>
      <InfiniteBanner clock={reverseClock.value}>
        <div className="flex space-x-6 pr-6">
          {bannerTwoImages.map((img, i) => (
            <Photo
              key={`set2-${i}`}
              src={img}
              onClick={(t) => setTarget((prev) => (prev !== t ? t : null))}
            />
          ))}
        </div>
      </InfiniteBanner>
    </div>
  )
}

export default Banners

"use client"

import {
  createContext,
  forwardRef,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react"
import { motion, useTransform } from "framer-motion"

import useIsomorphicLayoutEffect from "@/lib/hooks/useIsoLayoutEffect"

import * as utils from "./utils"

const CameraContext = createContext<utils.Camera | null>(null)

export const useCamera = () => {
  const camera = useContext(CameraContext)
  if (!camera) {
    throw new Error("useCamera can only be called inside of a Camera")
  }
  return camera
}

interface CameraProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Camera = ({ children, ...otherProps }: CameraProps) => {
  const [camera] = useState(() => new utils.Camera())
  const containerRef = useRef<HTMLDivElement | null>(null)
  const contentRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (camera) {
      // @ts-ignore
      camera.containerEl = containerRef.current
      // @ts-ignore
      camera.contentEl = contentRef.current
    }
  }, [])

  const translate = useTransform(
    [camera.motionValues.posX, camera.motionValues.posY],
    // @ts-ignore
    ([x, y]) => `${-x}px ${-y}px`
  )

  const transformOrigin = useTransform(
    [camera.motionValues.posX, camera.motionValues.posY],
    ([x, y]) => `calc(50% + ${x}px) calc(50% + ${y}px)`
  )

  console.log(camera.motionValues.zoom)

  return (
    <CameraContext.Provider value={camera}>
      {/* @ts-ignore */}
      <motion.div
        ref={containerRef}
        className="overflow-hidden"
        {...otherProps}
      >
        <motion.div
          className="w-full h-full"
          ref={contentRef}
          style={{
            translate,
            transformOrigin,
            scale: camera.motionValues.zoom,
            rotate: camera.motionValues.rotation,
          }}
        >
          {children}
        </motion.div>
      </motion.div>
    </CameraContext.Provider>
  )
}

// @ts-ignore
export interface CameraTargetProps
  extends React.HTMLAttributes<HTMLDivElement> {
  children: ((target: utils.CameraTarget) => React.ReactNode) | React.ReactNode
}

export const CameraTarget = forwardRef<utils.CameraTarget, CameraTargetProps>(
  ({ children, ...otherProps }, forwardedRef) => {
    const ref = useRef<HTMLDivElement>(null)
    const camera = useCamera()
    const [cameraTarget] = useState(() => new utils.CameraTarget(camera))

    useIsomorphicLayoutEffect(() => {
      // @ts-ignore
      cameraTarget.el = ref.current
      if (typeof forwardedRef === "function") {
        forwardedRef(cameraTarget)
      } else if (forwardedRef) {
        forwardedRef.current = cameraTarget
      }
    }, [])

    return (
      <div ref={ref} {...otherProps}>
        {typeof children === "function" ? children(cameraTarget) : children}
      </div>
    )
  }
)

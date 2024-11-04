"use client"

import React, { useEffect, useRef } from "react"
import { motion, useAnimation, useInView } from "framer-motion"

interface RevealAnimationProps {
  children: JSX.Element
  width?: "fit-content" | "100%"
  delay?: number
  slideTransition?: any // Custom transition for the slide
  contentTransition?: any // Transition for the content
  customSlideVariants?: { hidden: any; visible: any } // Custom slide variants
  customContentVariants?: { hidden: any; visible: any } // Custom content variants
  onEnterView?: () => void
  onExitView?: () => void
}

export const RevealAnimation = ({
  children,
  width = "100%",
  delay = 0,
  slideTransition = { duration: 0.4, ease: "easeIn" },
  contentTransition = { duration: 0.5, ease: "easeOut" },
  customSlideVariants,
  customContentVariants,
  onEnterView,
  onExitView,
}: RevealAnimationProps) => {
  const contentControls = useAnimation()
  const slideControls = useAnimation()
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true })

  useEffect(() => {
    if (isInView) {
      slideControls.start("visible")
      contentControls.start("visible")
      if (onEnterView) onEnterView()
    } else {
      slideControls.start("hidden")
      contentControls.start("hidden")
      if (onExitView) onExitView()
    }
  }, [isInView, contentControls, slideControls, onEnterView, onExitView])

  // Default slide variants
  const defaultSlideVariants = {
    // hidden: { scaleX: 1 },
    // visible: { scaleX: 0 },
    hidden: { left: 0 },
    visible: { left: "100%" },
  }

  // Default content variants
  const defaultContentVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 },
  }

  const slideVariants = customSlideVariants || defaultSlideVariants
  const contentVariants = customContentVariants || defaultContentVariants

  return (
    <div ref={ref} style={{ position: "relative", width, overflow: "hidden" }}>
      <motion.div
        variants={contentVariants}
        initial="hidden"
        animate={contentControls}
        transition={{ ...contentTransition, delay }}
      >
        {children}
      </motion.div>
      <motion.div
        className="absolute inset-0 bg-black dark:bg-neutral-100 z-[1000] rounded-l-xl"
        variants={slideVariants}
        initial="hidden"
        animate={slideControls}
        transition={slideTransition}
        style={{ originX: 0 }}
      />
    </div>
  )
}

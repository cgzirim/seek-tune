"use client"

import { ReactNode, useRef, useState } from "react"
import { motion } from "framer-motion"

import useIsomorphicLayoutEffect from "@/lib/hooks/useIsoLayoutEffect"

interface ScrollMountProps {
  withinTop?: string
  withinRight?: string
  withinBottom?: string
  withinLeft?: string
  once?: boolean
  children: ReactNode
}

const ScrollMount = ({
  withinTop = "0px",
  withinRight = "0px",
  withinBottom = "0px",
  withinLeft = "0px",
  once = false,
  children,
  ...otherProps
}: ScrollMountProps) => {
  const docRef = useRef<Element | null>(null)

  const [mounted, setMounted] = useState(false)

  useIsomorphicLayoutEffect(() => {
    if (docRef) {
      // @ts-ignore
      docRef.current = window.document
    }
  }, [])

  return (
    <motion.div
      {...otherProps}
      onViewportEnter={() => setMounted(true)}
      onViewportLeave={() => setMounted(false)}
      viewport={{
        root: docRef,
        margin: `${withinBottom} ${withinLeft} ${withinTop} ${withinRight}`,
        amount: "some",
        once,
      }}
    >
      {mounted && children}
    </motion.div>
  )
}

export default ScrollMount

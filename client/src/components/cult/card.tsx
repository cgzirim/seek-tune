"use client"

import { motion } from "framer-motion"

import { cn } from "@/lib/utils"

interface CardProps {
  text?: string
  description?: string
  children?: React.ReactNode
  gradient?: string
  image?: string
  price: string
  type: string
  textColor?: string
}

export function GradientCard({
  textColor = "text-white",
  gradient,
  text,
  price,
  type,
  description,
  children,
}: CardProps) {
  return (
    <motion.div
      className={cn(
        "relative  shadow-sm   h-[550px] w-full md:w-[400px] rounded-[28px]   border border-black/5",
        textColor
      )}
      style={{
        backgroundImage: gradient,
      }}
    >
      <div className="absolute inset-0 p-8 flex flex-col justify-between">
        <div className="flex flex-col space-y-3">
          <p className=" text-lg font-black leading-[1.2353641176] tracking-wide  ">
            {type}
          </p>

          <p className="mt-6 flex items-baseline justify-start gap-x-2">
            <span className="text-5xl font-bold tracking-tight ">${price}</span>

            <span className="text-sm font-semibold leading-6 tracking-wide text-neutral-600">
              USD
            </span>
          </p>

          <p
            className={cn(
              " text-lg font-semibold leading-[1.2353641176] tracking-wide text-neutral-200",
              textColor
            )}
          >
            {description}
          </p>
          <p
            className={cn(
              "text-3xl font-black tracking-[.007em]  mt-2",
              textColor
            )}
          >
            {text}
          </p>
        </div>
        {children}
      </div>
    </motion.div>
  )
}

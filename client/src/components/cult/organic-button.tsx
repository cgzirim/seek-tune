import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { ArrowUpRight } from "lucide-react"

import { cn } from "@/lib/utils"

const buttonContainerVariants = cva("", {
  variants: {
    size: {
      sm: "h-10",
      md: "h-12",
      lg: "h-14",
    },
    variantColor: {
      primary: "bg-neutral-950 text-white",
      secondary: "bg-secondary text-secondary-foreground",
    },
    animationSpeed: {
      slow: "duration-500",
      normal: "duration-400",
      fast: "duration-300",
    },
  },
  defaultVariants: {
    size: "md",
    animationSpeed: "normal",
  },
})

interface OrganicShapeButtonProps
  extends React.AnchorHTMLAttributes<HTMLAnchorElement>,
    VariantProps<typeof buttonContainerVariants> {
  asChild?: boolean
  label: string
  icon?: React.ReactNode
  animationSpeed?: "slow" | "normal" | "fast"
  leftShape?: React.ReactNode
  rightShape?: React.ReactNode
}

export const OrganicButton: React.FC<OrganicShapeButtonProps> =
  React.forwardRef<HTMLAnchorElement, OrganicShapeButtonProps>(
    (
      {
        asChild,
        label,
        className,
        icon = (
          <ArrowUpRight className=" group-hover:rotate-45 transition-all duration-100 group-hover:h-5 group-hover:w-5 h-4 w-4 stroke-white" />
        ),
        leftShape = <RoundedFlatRight />, // Accept leftShape as a prop
        rightShape = <FlatRoundedRight />, // Accept rightShape as a prop
        size = "md",
        variantColor = "primary",
        animationSpeed = "normal",
        ...props
      },
      ref
    ) => {
      const Comp = asChild ? Slot : "a" // Use Slot for composability or 'a' tag by default

      return (
        <div
          className={cn(
            buttonContainerVariants({
              size,
              animationSpeed,
            })
          )}
        >
          <Comp
            className={cn(
              "sm:w-btn-md focus:outline-none max-w-full disabled:cursor-not-allowed inline-block group",
              className
            )}
            {...props}
            ref={ref}
            target="_self"
          >
            <div className="relative flex grow">
              <div className="z-10 flex grow">
                <div className="h-full min-h-[40px] max-h-[40px] flex grow">
                  {/* SVG for the left part of the button */}

                  {leftShape}

                  {/* Button label */}
                  <div
                    className={cn(
                      buttonContainerVariants({ variantColor }),
                      "h-full truncate flex grow justify-between items-center"
                    )}
                  >
                    <span className="px-2  justify-between duration-400 flex w-full items-center transition-[padding] ease-in-out">
                      <span className="text-2xl font-brand font-semibold">
                        {label}
                      </span>
                    </span>
                  </div>

                  {/* SVG for the angle right part before the icon */}
                  <DescendingWall />
                </div>

                {/* Chevron icon container */}
                <div className="h-full min-h-[40px] max-h-[40px] flex grow -ml-1 ">
                  <AscendingWall />
                  <div
                    className={cn(
                      buttonContainerVariants({ variantColor }),
                      "h-full min-h-[40px] max-h-[40px] truncate flex grow justify-between items-center"
                    )}
                  >
                    <span className="px-0 group-hover:px-2 duration-400 flex items-center transition-all ease-in-out">
                      {icon}
                    </span>
                  </div>

                  {rightShape}
                </div>
              </div>
            </div>
          </Comp>
        </div>
      )
    }
  )

OrganicButton.displayName = "OrganicButton"

// Alternate Shapes, these are fun svgs to play around with for building unique button shapes

function AscendingWall() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -mr-[1px]"
    >
      <path
        d="M7.101 40H18V0H16C13.594 0 11.4403 1.49249 10.5955 3.74532L0.546698 30.5421C-1.1694 35.1184 2.21356 40 7.101 40Z"
        className="fill-neutral-950"
      ></path>
    </svg>
  )
}

function FlatRoundedRight() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -ml-[1px]"
    >
      <path
        d="M0 40V0H4C7.31371 0 10 2.68629 10 6V34C10 37.3137 7.31371 40 4 40H0Z"
        className="fill-neutral-950"
      ></path>
    </svg>
  )
}

function DescendingWall() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -ml-[1px]"
    >
      <path
        d="M10.899 0H0V40H2C4.40603 40 6.55968 38.5075 7.4045 36.2547L17.4533 9.45786C19.1694 4.88161 15.7864 0 10.899 0Z"
        className="fill-neutral-950"
      ></path>
    </svg>
  )
}

function RoundedFlatRight() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-cell-md max-h-cell-md -mr-[1px]"
    >
      <path
        d="M10 40V0H6C2.68629 0 0 2.68629 0 6V34C0 37.3137 2.68629 40 6 40H10Z"
        className="fill-neutral-950"
      ></path>
    </svg>
  )
}

function SmallPointRight() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -ml-[1px] fill-neutral-950"
    >
      {/* Unique design for the right filler SVG */}
      <path d="M0,0 L10,20 L0,40 Z" />
    </svg>
  )
}

// POINTY
function LargePointLeft() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -mr-[1px] fill-neutral-950"
    >
      {/* Unique design for the middle-right SVG */}
      <path d="M0,20 L17,0 L17,40 L0,20 Z" />
    </svg>
  )
}

function SmallSoftPointRight() {
  return (
    <svg
      viewBox="0 0 38 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] fill-neutral-950 -mr-[1px]"
    >
      {/* Softened design for a right-pointing caret */}
      <path
        d="M0,0 
             Q0,20 0,40 
             L8,39 
             Q48,20 8,0 
             L0,0 Z"
      />
    </svg>
  )
}

function SmallSoftPointLeft() {
  return (
    <svg
      viewBox="0 0 38 40" // Adjusted viewBox width to keep the design within the 40px height
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] fill-neutral-950"
    >
      {/* Softened design for a left-pointing caret */}
      <path
        d="M38,0 
             Q38,20 38,40 
             L30,39 
             Q-1,20 30,0 
             L38,0 Z"
      />
    </svg>
  )
}

function LargeSoftPointLeft() {
  return (
    <svg
      viewBox="0 0 38 40" // Adjusted viewBox width to keep the design within the 40px height
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] fill-neutral-950"
    >
      {/* Softened design for a left-pointing caret */}
      <path
        d="M38,0 
             Q38,20 38,40 
             L30,39 
             Q-10,20 30,0 
             L38,0 Z"
      />
    </svg>
  )
}

function SmallLeftCrater() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] fill-neutral-950"
    >
      {/* Softened design for the right filler SVG */}
      <path d="M0,0 Q10,20 0,40 L10,40 Q10,20 10,0 L0,0 Z" />
    </svg>
  )
}

function LargePointRight() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -mr-[1px] fill-neutral-950"
    >
      {/* Adjusted design for the middle-right angle SVG to point the other way */}
      <path d="M0,0 L17,20 L0,40 V0" />
    </svg>
  )
}

function StraightLineSvg() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full -mr-[1px] fill-neutral-950"
    >
      {/* Unique design for the left front SVG */}
      <path d="M0,0 L10,0 L10,40 L0,40 Z" />
    </svg>
  )
}

// BUBBLE SET
function LargeBubbleLeft() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -mr-[1px] fill-neutral-950"
    >
      <path d="M18,0 Q0,20 18,40 V0" />
    </svg>
  )
}

function SmallBubbleRight() {
  return (
    <svg
      viewBox="0 0 10 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -ml-[1px] fill-neutral-950"
    >
      <path d="M10,0 Q10,20 10,40 L0,40 L0,0 Z" />
    </svg>
  )
}

function LargeBubbleRight() {
  return (
    <svg
      viewBox="0 0 18 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px] -mr-[1px] fill-neutral-950"
    >
      <path d="M0,0 Q18,20 0,40 V0" />
    </svg>
  )
}

function FlatLeftCircleSVG() {
  return (
    <svg
      viewBox="0 0 43 40" // Keep viewBox the same to match the height
      xmlns="http://www.w3.org/2000/svg"
      className="h-full fill-neutral-950"
    >
      <path
        d="M 43,20
             A 20,20 0 0 0 23,0
             L 0,0
             L 0,40
             L 23,40
             A 20,20 0 0 0 43,20
             Z"
      />
    </svg>
  )
}

function FlatRightCircleSVG() {
  return (
    <svg
      viewBox="0 0 43 40" // Set the viewBox width to 43 and height to 40 to match the button's height
      xmlns="http://www.w3.org/2000/svg"
      className="h-full fill-neutral-950"
    >
      <path
        d="M 0,20
             A 20,20 0 0 1 20,0
             L 43,0
             L 43,40
             L 20,40
             A 20,20 0 0 1 0,20
             Z"
      />
    </svg>
  )
}

function CircleSVG() {
  return (
    <svg
      width="40" // Width of the SVG
      height="40" // Height of the SVG
      viewBox="0 0 40 40" // ViewBox to ensure the circle is fully visible
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-full min-h-[40px] max-h-[40px]  -ml-[1px] fill-neutral-950"
    >
      <circle
        cx="20" // X-coordinate of the circle center
        cy="20" // Y-coordinate of the circle center
        r="15" // Radius of the circle
      />
    </svg>
  )
}

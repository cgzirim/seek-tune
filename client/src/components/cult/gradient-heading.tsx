import React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const headingVariants = cva(
  "tracking-tight pb-3 bg-clip-text font-sans text-transparent",
  {
    variants: {
      variant: {
        default:
          " bg-gradient-to-t from-neutral-800 to-neutral-950 dark:from-stone-200 dark:to-neutral-200",
        brand:
          " bg-gradient-to-r from-white via-yellow-100 to-black dark:from-stone-200 dark:to-neutral-200",
        light: "bg-gradient-to-t from-neutral-50 to-neutral-100",
        lightSecondary: "bg-gradient-to-t from-neutral-200 to-neutral-300",
        secondary:
          "bg-gradient-to-t  from-primary-foreground to-muted-foreground",
      },
      size: {
        default: "text-2xl sm:text-3xl lg:text-4xl",
        xxs: "text-base sm:text-lg lg:text-lg",
        xs: "text-lg sm:text-xl lg:text-2xl",
        sm: "text-xl sm:text-2xl lg:text-3xl",
        md: "text-2xl sm:text-3xl lg:text-4xl",
        lg: "text-3xl sm:text-4xl lg:text-5xl",
        xl: "text-4xl sm:text-5xl lg:text-6xl",
        xxl: "text-5xl sm:text-6xl lg:text-[6rem]",
        xxxl: "text-5xl sm:text-6xl lg:text-[19rem]",
      },
      weight: {
        default: "font-bold",
        thin: "font-thin",
        base: "font-base",
        semi: "font-semibold",
        bold: "font-bold",
        black: "font-black",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
      weight: "default",
    },
  }
)

export interface HeadingProps extends VariantProps<typeof headingVariants> {
  asChild?: boolean
  children: React.ReactNode
  className?: string
}

const GradientHeading = React.forwardRef<HTMLHeadingElement, HeadingProps>(
  ({ asChild, variant, weight, size, className, children, ...props }, ref) => {
    const Comp = asChild ? Slot : "h3" // default to 'h3' if not a child
    return (
      <Comp ref={ref} {...props} className={className}>
        <span className={cn(headingVariants({ variant, size, weight }))}>
          {children}
        </span>
      </Comp>
    )
  }
)

GradientHeading.displayName = "GradientHeading"

export { GradientHeading, headingVariants }

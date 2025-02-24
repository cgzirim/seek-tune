import Image from "next/image"

import { cn } from "@/lib/utils"

export const BentoGrid = ({
  className,
  children,
}: {
  className?: string
  children?: React.ReactNode
}) => {
  return (
    <div
      className={cn(
        "grid md:auto-rows-[18rem] grid-cols-1 md:grid-cols-3 gap-4 max-w-7xl mx-auto ",
        className
      )}
    >
      {children}
    </div>
  )
}

export function BentoImageCard({
  text,
  description,
  image,
  children,
  className,
}: any) {
  return (
    <div
      className={cn(
        "row-span-1  group/bento hover:shadow-xl transition duration-200 shadow-input relative flex-shrink-0 fit-content bg-none max-w-md shadow-sm  rounded-[28px]  border border-black/5 group",
        className
      )}
    >
      <Image
        src={image}
        alt={`Slide ${text}`}
        width={405}
        height={704}
        className="w-full h-full object-cover rounded-[28px] border border-black/10"
      />

      <div className="absolute inset-0 p-8 flex flex-col justify-between">
        <div>
          <p className=" text-lg font-semibold leading-[1.2353641176] tracking-wide text-black">
            {description}
          </p>
          <p className="text-3xl font-black tracking-[.007em] text-black mt-2">
            {text}
          </p>
        </div>
        {children}
      </div>
    </div>
  )
}

export const BentoGridItem = ({
  className,
  title,
  description,
  header,
  icon,
}: {
  className?: string
  title?: string | React.ReactNode
  description?: string | React.ReactNode
  header?: React.ReactNode
  icon?: React.ReactNode
}) => {
  return (
    <div
      className={cn(
        "row-span-1  group/bento hover:shadow-xl transition duration-200 shadow-input dark:shadow-none p-4 dark:bg-black dark:border-white/[0.2] bg-orange-50/20  justify-between flex flex-col space-y-4 rounded-[28px] border border-black/5",
        className
      )}
    >
      {header}
      <div className="group-hover/bento:translate-x-2 transition duration-200">
        {icon}
        <div className="font-sans font-bold text-indigo-900 dark:text-indigo-200 mb-2 mt-2">
          {title}
        </div>
        <div className="font-sans font-normal text-indigo-900 text-xs dark:text-indigo-300">
          {description}
        </div>
      </div>
    </div>
  )
}

export const BentoGridItemCta = ({
  className,
  title,
  description,
  header,
  icon,
}: {
  className?: string
  title?: string | React.ReactNode
  description?: string | React.ReactNode
  header?: React.ReactNode
  icon?: React.ReactNode
}) => {
  return (
    <div
      className={cn(
        "row-span-1  group/bento hover:shadow-xl transition duration-200 shadow-input dark:shadow-none p-4 bg-black  justify-between flex flex-col space-y-4 rounded-[28px] border border-black/5",
        className
      )}
    >
      {header}
      <div className="group-hover/bento:translate-x-2 transition duration-200">
        {icon}
        <div className="font-sans font-bold  text-orange-100 mb-2 mt-2">
          {title}
        </div>
        <div className="font-sans font-normal  text-xs text-neutral-100">
          {description}
        </div>
      </div>
    </div>
  )
}

"use client"

import { useCallback, useState } from "react"
import { motion } from "framer-motion"

export function Navbar({ activeSection }) {
  const [activeTab, setActiveTab] = useState("hero")

  const handleTabClick = useCallback((id) => {
    const section = document.querySelector(`#${id}`)
    if (section) {
      section.scrollIntoView({ behavior: "smooth" })
      setActiveTab(id)
    }
  }, [])

  // Dynamically generate tabs with icons and conditional styling
  const tabs = [
    {
      id: "hero",
      label: "H",
      icon: (
        <CultIcon
          className={`h-6 w-6 ${
            activeSection === "hero" ? "text-orange-100" : "text-neutral-100/60"
          }`}
        />
      ),
    },
    {
      id: "feature",
      label: "Code",
      icon: (
        <CodeIcon
          className={`h-6 w-6 ${
            activeSection === "feature"
              ? "text-orange-200"
              : "text-neutral-100/60"
          }`}
        />
      ),
    },
    {
      id: "testimonial",
      label: "Ship",
      icon: (
        <RocketIcon
          className={`h-6 w-6 ${
            activeSection === "testimonial"
              ? "text-orange-300"
              : "text-neutral-100/60"
          }`}
        />
      ),
    },
    {
      id: "price",
      label: "start",
      icon: (
        <DollarIcon
          className={`h-6 w-6 ${
            activeSection === "price"
              ? "text-orange-400"
              : "text-neutral-100/60"
          }`}
        />
      ),
    },
  ]

  return (
    <div className="flex space-x-4 sticky top-0 z-50 bg-black/60 px-1 py-[3px] rounded-full border border-black">
      <ul className="flex w-full justify-between">
        {tabs.map((tab) => {
          const isActive = activeTab === tab.id || activeSection === tab.id

          return (
            <motion.button
              key={tab.id}
              onClick={() => handleTabClick(tab.id)}
              className="relative flex items-center justify-center px-4 py-2 text-lg cursor-pointer font-medium outline-none transition focus-visible:outline-2"
              style={{ WebkitTapHighlightColor: "transparent" }}
            >
              {isActive && (
                <motion.div
                  layoutId="highlight"
                  className="absolute inset-0 bg-black mix-blend-difference"
                  style={{ borderRadius: 9999 }}
                  transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                />
              )}

              <div className="z-20 flex items-center">{tab.icon}</div>
            </motion.button>
          )
        })}
      </ul>
    </div>
  )
}

function CultIcon(props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M8 17H16M11.0177 2.764L4.23539 8.03912C3.78202 8.39175 3.55534 8.56806 3.39203 8.78886C3.24737 8.98444 3.1396 9.20478 3.07403 9.43905C3 9.70352 3 9.9907 3 10.5651V17.8C3 18.9201 3 19.4801 3.21799 19.908C3.40973 20.2843 3.71569 20.5903 4.09202 20.782C4.51984 21 5.07989 21 6.2 21H17.8C18.9201 21 19.4802 21 19.908 20.782C20.2843 20.5903 20.5903 20.2843 20.782 19.908C21 19.4801 21 18.9201 21 17.8V10.5651C21 9.9907 21 9.70352 20.926 9.43905C20.8604 9.20478 20.7526 8.98444 20.608 8.78886C20.4447 8.56806 20.218 8.39175 19.7646 8.03913L12.9823 2.764C12.631 2.49075 12.4553 2.35412 12.2613 2.3016C12.0902 2.25526 11.9098 2.25526 11.7387 2.3016C11.5447 2.35412 11.369 2.49075 11.0177 2.764Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}
function RocketIcon(props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M12 14.9998L9 11.9998M12 14.9998C13.3968 14.4685 14.7369 13.7985 16 12.9998M12 14.9998V19.9998C12 19.9998 15.03 19.4498 16 17.9998C17.08 16.3798 16 12.9998 16 12.9998M9 11.9998C9.53214 10.6192 10.2022 9.29582 11 8.04976C12.1652 6.18675 13.7876 4.65281 15.713 3.59385C17.6384 2.53489 19.8027 1.98613 22 1.99976C22 4.71976 21.22 9.49976 16 12.9998M9 11.9998H4C4 11.9998 4.55 8.96976 6 7.99976C7.62 6.91976 11 7.99976 11 7.99976M4.5 16.4998C3 17.7598 2.5 21.4998 2.5 21.4998C2.5 21.4998 6.24 20.9998 7.5 19.4998C8.21 18.6598 8.2 17.3698 7.41 16.5898C7.02131 16.2188 6.50929 16.0044 5.97223 15.9878C5.43516 15.9712 4.91088 16.1535 4.5 16.4998Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}
function CodeIcon(props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M7 15L10 12L7 9M13 15H17M7.8 21H16.2C17.8802 21 18.7202 21 19.362 20.673C19.9265 20.3854 20.3854 19.9265 20.673 19.362C21 18.7202 21 17.8802 21 16.2V7.8C21 6.11984 21 5.27976 20.673 4.63803C20.3854 4.07354 19.9265 3.6146 19.362 3.32698C18.7202 3 17.8802 3 16.2 3H7.8C6.11984 3 5.27976 3 4.63803 3.32698C4.07354 3.6146 3.6146 4.07354 3.32698 4.63803C3 5.27976 3 6.11984 3 7.8V16.2C3 17.8802 3 18.7202 3.32698 19.362C3.6146 19.9265 4.07354 20.3854 4.63803 20.673C5.27976 21 6.11984 21 7.8 21Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}
function DollarIcon(props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M6 16C6 18.2091 7.79086 20 10 20H14C16.2091 20 18 18.2091 18 16C18 13.7909 16.2091 12 14 12H10C7.79086 12 6 10.2091 6 8C6 5.79086 7.79086 4 10 4H14C16.2091 4 18 5.79086 18 8M12 2V22"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

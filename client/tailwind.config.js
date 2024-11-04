/** @type {import('tailwindcss').Config} */
const defaultTheme = require("tailwindcss/defaultTheme")
module.exports = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./app/**/*.{ts,tsx}",
    "./src/**/*.{ts,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },

        yellow: {
          50: "#fff5eb",
          100: "#ffedd6",
          200: "#ffdba8",
          300: "#ffc266",
          400: "#feaf39",
          500: "#fb9e13",
          600: "#ec8d09",
          700: "#c37509",
          800: "#9d6510",
          900: "#7e5810",
          950: "#463106",
        },

        indigo: {
          50: "#ebeeff",
          100: "#dbdfff",
          200: "#bfc4ff",
          300: "#989eff",
          400: "#746fff",
          500: "#624eff",
          600: "#562efd",
          700: "#4b22e0",
          800: "#3c1fb4",
          900: "#221C35",
          950: "#0d0821",
        },

        orange: {
          50: "#fff2eb",
          100: "#ffe4d1",
          200: "#ffcba3",
          300: "#ffab6b",
          400: "#ff8533",
          500: "#ff6c0a",
          600: "#ff5900",
          700: "#ca4502",
          800: "#a23e0b",
          900: "#83380c",
          950: "#481d04",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
        spotlight: {
          "0%": {
            opacity: 0,
            transform: "translate(-72%, -62%) scale(0.5)",
          },
          "100%": {
            opacity: 1,
            transform: "translate(-50%,-40%) scale(1)",
          },
        },
        spin: {
          from: {
            transform: "rotate(0deg)",
          },
          to: {
            transform: "rotate(360deg)",
          },
        },
        rotateIcon: {
          "0%": { transform: "rotate(0deg)" },
          "100%": { transform: "rotate(-90deg)" },
        },
        rotateIconDown: {
          "0%": { transform: "rotate(180deg)" },
          "100%": { transform: "rotate(0deg)" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        spotlight: "spotlight 2s ease .75s 1 forwards",
        "spin-slow": "spin 79s linear infinite",
        "spin-medium": "spin 9s linear infinite",
        "rotate-once-2s-delay": "rotateIcon 1s linear 2.1s forwards", // 1s for animation duration, 2s for delay, 1 for running once
        "rotate-once-down-2s-delay": "rotateIconDown 2s ease .75s forwards", // 1s for animation duration, 2s for delay, 1 for running once
      },
      fontFamily: {
        sans: ["var(--font-satoshi)"],
        brand: ["var(--font-brand)"],
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}

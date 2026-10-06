/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./ui/**/*.templ",
    "./ui/**/*.html",
    "./ui/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "rgb(var(--color-primary) / <alpha-value>)",
          dark: "rgb(var(--color-primary-dark) / <alpha-value>)",
          light: "rgb(var(--color-primary-light) / <alpha-value>)",
        },
        accent: {
          DEFAULT: "rgb(var(--color-accent) / <alpha-value>)",
          hover: "rgb(var(--color-accent-hover) / <alpha-value>)",
          light: "rgb(var(--color-accent-light) / <alpha-value>)",
        },
        surface: {
          DEFAULT: "rgb(var(--color-surface) / <alpha-value>)",
          muted: "rgb(var(--color-surface-muted) / <alpha-value>)",
        },
        content: {
          DEFAULT: "rgb(var(--color-content) / <alpha-value>)",
          muted: "rgb(var(--color-content-muted) / <alpha-value>)",
        },
        // Legacy colors mapped to semantic tokens for smooth transition
        legal: {
          burgundy: "rgb(var(--color-primary) / <alpha-value>)",
          burgundydark: "rgb(var(--color-primary-dark) / <alpha-value>)",
          burgundylight: "rgb(var(--color-primary-light) / <alpha-value>)",
          gold: "rgb(var(--color-accent) / <alpha-value>)",
          goldhover: "rgb(var(--color-accent-hover) / <alpha-value>)",
          goldlight: "rgb(var(--color-accent-light) / <alpha-value>)",
          light: "rgb(var(--color-surface) / <alpha-value>)",
          muted: "rgb(var(--color-content-muted) / <alpha-value>)",
        }
      }
    },
  },
  plugins: [],
}

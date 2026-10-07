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
          DEFAULT: "color-mix(in srgb, var(--color-primary) calc(100% * <alpha-value>), transparent)",
          dark: "color-mix(in srgb, var(--color-primary-dark) calc(100% * <alpha-value>), transparent)",
          light: "color-mix(in srgb, var(--color-primary-light) calc(100% * <alpha-value>), transparent)",
        },
        accent: {
          DEFAULT: "color-mix(in srgb, var(--color-accent) calc(100% * <alpha-value>), transparent)",
          hover: "color-mix(in srgb, var(--color-accent-hover) calc(100% * <alpha-value>), transparent)",
          light: "color-mix(in srgb, var(--color-accent-light) calc(100% * <alpha-value>), transparent)",
        },
        surface: {
          DEFAULT: "color-mix(in srgb, var(--color-surface) calc(100% * <alpha-value>), transparent)",
          muted: "color-mix(in srgb, var(--color-surface-muted) calc(100% * <alpha-value>), transparent)",
        },
        content: {
          DEFAULT: "color-mix(in srgb, var(--color-content) calc(100% * <alpha-value>), transparent)",
          muted: "color-mix(in srgb, var(--color-content-muted) calc(100% * <alpha-value>), transparent)",
        },
        // Legacy colors mapped to semantic tokens for smooth transition
        legal: {
          burgundy: "color-mix(in srgb, var(--color-primary) calc(100% * <alpha-value>), transparent)",
          burgundydark: "color-mix(in srgb, var(--color-primary-dark) calc(100% * <alpha-value>), transparent)",
          burgundylight: "color-mix(in srgb, var(--color-primary-light) calc(100% * <alpha-value>), transparent)",
          gold: "color-mix(in srgb, var(--color-accent) calc(100% * <alpha-value>), transparent)",
          goldhover: "color-mix(in srgb, var(--color-accent-hover) calc(100% * <alpha-value>), transparent)",
          goldlight: "color-mix(in srgb, var(--color-accent-light) calc(100% * <alpha-value>), transparent)",
          light: "color-mix(in srgb, var(--color-surface) calc(100% * <alpha-value>), transparent)",
          muted: "color-mix(in srgb, var(--color-content-muted) calc(100% * <alpha-value>), transparent)",
        }
      },
      fontFamily: {
        condensed: ['"Bebas Neue"', '"Barlow Condensed"', 'Impact', 'sans-serif'],
        'serif-title': ['"Playfair Display"', 'Georgia', 'serif'],
      }
    },
  },
  plugins: [],
}

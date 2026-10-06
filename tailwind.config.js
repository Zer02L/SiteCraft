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
        legal: {
          burgundy: "#7A1C24",
          burgundydark: "#521217",
          burgundylight: "#982832",
          gold: "#D4AF37",
          goldhover: "#B8972E",
          goldlight: "#FDF8E8",
          light: "#FDFBF7",
          muted: "#716565"
        }
      }
    },
  },
  plugins: [],
}

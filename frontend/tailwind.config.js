/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        main: {
          100: "#e6f7f9",
          200: "#9ce0e5",
          300: "#6bd0d9",
          400: "#39c1cc",
          500: "#08b1bf",
          600: "#068e99",
          700: "#056a73",
          800: "#03474c",
          900: "#022326",
        },
      },
    },
    screens: {
      sD: { max: "1344px" },

      lt: { max: "1200px" },

      tt: { max: "944px" },

      st: { max: "704px" },

      mp: { max: "544px" },
    },
  },
  plugins: [],
};

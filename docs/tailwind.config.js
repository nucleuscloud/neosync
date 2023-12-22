/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{js,jsx,ts,tsx}', './docs/*.mdx'],
  theme: {
    extend: {
      container: {
        center: true,
        padding: '2rem',
        // screens: {
        //   xl: '2000px',
        //   '2xl': '1400px',
        // },
      },
    },
  },
  darkMode: ['class', '[data-theme="dark"]'],
  plugins: [],
};

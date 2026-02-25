module.exports = {
  darkMode: 'class', // 关键：支持通过 class="dark" 切换主题
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#121212',
          surface: '#1e1e1e',
          border: '#333333',
          text: '#e0e0e0'
        }
      }
    },
  },
  plugins: [],
}

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: '#18181b',
        secondary: '#27272a',
        cta: '#a3e635',
        'cta-hover': '#bef264',
        background: '#09090b',
        surface: '#18181b',
        'surface-hover': '#27272a',
        text: '#fafafa',
        'text-secondary': '#d4d4d8',
        'text-muted': '#71717a',
        border: '#27272a',
        'border-subtle': '#1f1f23',
        'accent-blue': '#60a5fa',
        'accent-amber': '#fbbf24',
        'accent-rose': '#fb7185',
        'accent-violet': '#a78bfa',
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
        serif: ['Newsreader', 'Georgia', 'serif'],
      },
    },
  },
  plugins: [],
}

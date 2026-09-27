/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Design tokens mirrored from the reference index.html.
        page: '#EDF0F5',
        surface: {
          DEFAULT: '#FFFFFF',
          muted: '#F6F7F9',
          sunken: '#F1F3F6',
        },
        border: {
          DEFAULT: '#E3E7ED',
          strong: '#D3D9E2',
        },
        ink: {
          strong: '#0F172A',
          DEFAULT: '#334155',
          muted: '#64748B',
          faint: '#94A3B8',
        },
        accent: {
          DEFAULT: '#2563EB',
          hover: '#1D4ED8',
          50: '#EFF5FF',
          100: '#DCE8FE',
          200: '#BED3FC',
          text: '#1E4FD1',
        },
        warn: { bg: '#FFF8EB', border: '#F8DDA6', icon: '#D97706', title: '#8A4B08' },
        danger: { bg: '#FFF4F4', border: '#F6C9C9', icon: '#DC2626', title: '#9B1C1C' },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'Segoe UI', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'SF Mono', 'Menlo', 'monospace'],
      },
      borderRadius: {
        sm: '6px',
        DEFAULT: '10px',
        lg: '16px',
      },
      boxShadow: {
        card: '0 1px 2px rgba(15,23,42,.04)',
        shell: '0 1px 2px rgba(15,23,42,.04), 0 12px 40px -12px rgba(15,23,42,.10)',
        pop: '0 20px 40px rgba(15,23,42,.18)',
      },
    },
  },
  plugins: [],
}

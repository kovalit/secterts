import { defineConfig } from 'vite'
import { crx, type ManifestV3Export } from '@crxjs/vite-plugin'
import manifest from './manifest.json'

// CRXJS reads manifest.json, bundles the popup/options pages, the service
// worker and copies the icons from public/ into dist/.
export default defineConfig({
  plugins: [crx({ manifest: manifest as ManifestV3Export })],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5174,
    strictPort: true,
  },
})

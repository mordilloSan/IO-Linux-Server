import { defineConfig } from "vite";

import tsconfigPaths from "vite-tsconfig-paths";
import react from "@vitejs/plugin-react-swc";
import svgr from "@svgr/rollup";

// https://vitejs.dev/config/
export default defineConfig({
  base: "/",
  clearScreen: false,
  plugins: [react(), svgr(), tsconfigPaths()],
  build: {
    chunkSizeWarningLimit: 2000,
    outDir: "../go-backend/frontend",
    emptyOutDir: true,
  },
  resolve: {
    conditions: ["mui-modern", "module", "browser", "development|production"],
  },
});

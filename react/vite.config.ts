import { defineConfig } from "vite";

import tsconfigPaths from "vite-tsconfig-paths";
import react from "@vitejs/plugin-react-swc";
import svgr from "@svgr/rollup";

export default defineConfig({
  base: "/",
  clearScreen: false,
  plugins: [react(), svgr(), tsconfigPaths()],
  build: {
    target: "es2017",
    chunkSizeWarningLimit: 2000,
    outDir: "../go-backend/frontend",
    emptyOutDir: true,
    minify: "esbuild",
  },
  resolve: {
    conditions: ["mui-modern", "module", "browser", "development|production"],
  },
});

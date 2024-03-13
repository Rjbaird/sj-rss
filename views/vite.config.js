import { resolve } from "path";
import { defineConfig } from "vite";

export default defineConfig({
  build: {
    lib: {
      entry: [
        resolve(__dirname, "/js/htmx.js"),
        resolve(__dirname, "/js/alpine.js"),
      ],
      formats: ["es"],
      name: "[name]",
      fileName: "[name]",
    },
    outDir: resolve(__dirname, "static"),
    emptyOutDir: false,
  },
});

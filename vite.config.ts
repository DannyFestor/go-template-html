import tailwindcss from "@tailwindcss/vite";
import { defineConfig, type Plugin } from "vite-plus";

const assetDir = "internal/asset";

// Vite empties the output directory before each build, and the committed .gitkeep lets a fresh clone embed it.
const keepOutputDirInGit: Plugin = {
  name: "keep-output-dir-in-git",
  generateBundle() {
    this.emitFile({ type: "asset", fileName: ".gitkeep", source: "" });
  },
};

export default defineConfig({
  plugins: [tailwindcss(), keepOutputDirInGit],
  build: {
    outDir: `${assetDir}/dist`,
    emptyOutDir: true,
    // go:embed skips the default .vite/ directory.
    manifest: "manifest.json",
    // The strict CSP allows neither the inline polyfill nor data: URIs.
    modulePreload: { polyfill: false },
    assetsInlineLimit: 0,
    rolldownOptions: {
      input: `${assetDir}/src/app.ts`,
    },
  },
  test: {
    environment: "happy-dom",
    include: [`${assetDir}/src/**/*.test.ts`],
  },
  // The closest oxlint equivalents of funlen, cyclop, nestif and argument-limit in .golangci.yml.
  lint: {
    env: { browser: true },
    categories: { correctness: "error", suspicious: "error", perf: "error" },
    rules: {
      "max-lines-per-function": ["error", { max: 40 }],
      complexity: ["error", { max: 10 }],
      "max-depth": ["error", { max: 4 }],
      "max-params": ["error", { max: 5 }],
    },
  },
  // Markdown keeps its hand-aligned tables, and TOML is left to the tools that own each file.
  fmt: {
    ignorePatterns: ["**/*.md", "**/*.toml"],
  },
});

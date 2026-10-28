/// <reference types="vitest/config" />
import { defineConfig, loadEnv } from "vite"
import react from "@vitejs/plugin-react"

// Feature code reads configuration through `process.env.REACT_APP_*` and
// `process.env.NODE_ENV` (Create React App conventions). In the browser
// bundle, replace `process.env` with a static object holding those values so
// unset variables read as `undefined` instead of throwing. Under Vitest the
// real Node `process.env` is used, so specs can still set variables at runtime.
export default defineConfig(({ mode }) => {
  const isTest = mode === "test" || process.env.VITEST !== undefined
  const fileEnv = loadEnv(mode, process.cwd(), "REACT_APP_")
  const env: Record<string, string> = {
    NODE_ENV: mode === "production" ? "production" : "development",
  }
  for (const [key, value] of Object.entries({ ...fileEnv, ...process.env })) {
    if (key.startsWith("REACT_APP_") && value !== undefined) env[key] = value
  }

  return {
    plugins: [react()],
    define: isTest ? {} : { "process.env": JSON.stringify(env) },
    server: { port: 3000 },
    preview: { port: 3000 },
    build: { outDir: "build" },
    test: {
      environment: "jsdom",
      globals: true,
      setupFiles: ["./src/setupTests.ts"],
      include: ["src/**/__tests__/**/*.spec.{ts,tsx}"],
    },
  }
})

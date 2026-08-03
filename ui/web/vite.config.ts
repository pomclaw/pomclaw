import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const protocol = env.VITE_BACKEND_PROTOCOL || "http";
  const backendPort = env.VITE_BACKEND_PORT || "9600";
  const backendHost = env.VITE_BACKEND_HOST || "localhost";

  // 标准端口不显示在 URL 中
  const isStandardPort =
    (protocol === "https" && backendPort === "443") ||
    (protocol === "http" && backendPort === "80");

  const backendTarget = isStandardPort
    ? `${protocol}://${backendHost}`
    : `${protocol}://${backendHost}:${backendPort}`;

  return {
    base: env.VITE_PUBLIC_PATH,
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      port: 5173,
      proxy: {
        "/pomclaw-api": { target: backendTarget, changeOrigin: true, ws: true, timeout: 30000 },
        "/health": { target: backendTarget, changeOrigin: true },
      },
    },
    build: {
      outDir: "dist",
      emptyOutDir: true,
    },
  };
});

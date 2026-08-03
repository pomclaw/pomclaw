import { HttpClient } from "@/api/http-client";

const _protocol = import.meta.env.VITE_BACKEND_PROTOCOL || "http";
const _host = import.meta.env.VITE_BACKEND_HOST;
const _port = import.meta.env.VITE_BACKEND_PORT;

// 生产环境：使用完整 URL；开发环境：使用相对路径（走 Vite 代理）
export const API_BASE_URL =
  _host && !import.meta.env.DEV
    ? `${_protocol}://${_host}${_port && !isStandardPort(_protocol, _port) ? `:${_port}` : ""}`
    : "";

// 检查是否是标准端口
function isStandardPort(protocol: string, port: string): boolean {
  return (protocol === "https" && port === "443") || (protocol === "http" && port === "80");
}

export const API_PREFIX = import.meta.env.VITE_API_PREFIX || "";

export const WS_URL =
  import.meta.env.VITE_WS_URL || (API_PREFIX ? `${API_PREFIX}/ws` : "/ws");

// 预认证客户端：用于登录/注册等无需 token 的请求，URL 构建逻辑与主客户端一致
export const preAuthClient = new HttpClient(
  API_BASE_URL,
  () => "",
  () => "",
  () => "",
  API_PREFIX,
);

import { useState } from "react";
import { useNavigate, useLocation } from "react-router";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/use-auth-store";
import { ROUTES } from "@/lib/constants";
import { preAuthClient } from "@/lib/api-config";
import { LoginLayout } from "./login-layout";
import { LoginForm } from "./login-form";
import { RegisterForm } from "./register-form";

interface AuthResponse {
  user_id: string;
  access_token: string;
  refresh_token: string;
  username: string;
  expires_in: number;
  token_type: string;
}

export function LoginPage() {
  const { t } = useTranslation("login");
  const [showRegister, setShowRegister] = useState(false);

  const setCredentials = useAuthStore((s) => s.setCredentials);
  const navigate = useNavigate();
  const location = useLocation();

  const from =
    (location.state as { from?: { pathname: string } })?.from?.pathname ??
    ROUTES.OVERVIEW;

  async function handleLogin(username: string, password: string) {
    const data = await preAuthClient.post<AuthResponse>("/v1/auth/login", { username, password });
    setCredentials(data.access_token, data.user_id, data.username, "password");
    useAuthStore.getState().setTenantSelected(true);
    navigate(from, { replace: true });
  }

  async function handleRegister(email: string, username: string, password: string) {
    await preAuthClient.post("/v1/auth/register", { email, username, password });
    setTimeout(() => setShowRegister(false), 200);
  }

  return (
    <LoginLayout subtitle={t("subtitle")}>
      <div className="mt-6 space-y-4">
        {showRegister ? (
          <>
            <RegisterForm onSubmit={handleRegister} />
            <button
              type="button"
              onClick={() => setShowRegister(false)}
              className="w-full text-xs text-primary hover:underline py-2"
            >
              已有账户？返回登录
            </button>
          </>
        ) : (
          <>
            <LoginForm onSubmit={handleLogin} />
            <button
              type="button"
              onClick={() => setShowRegister(true)}
              className="w-full text-xs text-primary hover:underline py-2"
            >
              还没有账户？立即注册
            </button>
          </>
        )}
      </div>
    </LoginLayout>
  );
}

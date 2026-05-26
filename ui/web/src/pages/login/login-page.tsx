import { useState } from "react";
import { useNavigate, useLocation } from "react-router";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/use-auth-store";
import { ROUTES } from "@/lib/constants";
import { LoginLayout } from "./login-layout";
import { LoginTabs, type LoginMode } from "./login-tabs";
import { LoginForm } from "./login-form";
import { RegisterForm } from "./register-form";

interface AuthResponse {
  user_id: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export function LoginPage() {
  const { t } = useTranslation("login");
  const [mode, setMode] = useState<LoginMode>("login");

  const setCredentials = useAuthStore((s) => s.setCredentials);
  const navigate = useNavigate();
  const location = useLocation();

  const from =
    (location.state as { from?: { pathname: string } })?.from?.pathname ??
    ROUTES.OVERVIEW;

  async function handleLogin(username: string, password: string) {
    const res = await fetch("/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: "Login failed" }));
      throw new Error(error.error?.message || error.error || "Login failed");
    }

    const data: AuthResponse = await res.json();

    // Store token and username as userId
    setCredentials(data.access_token, data.user_id);

    // Set tenant as selected (single tenant mode, no multi-tenant)
    useAuthStore.getState().setTenantSelected(true);

    navigate(from, { replace: true });
  }

  async function handleRegister(email: string, username: string, password: string) {
    const res = await fetch("/v1/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, username, password }),
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: "Registration failed" }));
      throw new Error(error.error?.message || error.error || "Registration failed");
    }

    // After successful registration, switch to login tab
    setTimeout(() => setMode("login"), 2000);
  }

  return (
    <LoginLayout subtitle={t("subtitle")}>
      <LoginTabs mode={mode} onModeChange={setMode} />
      {mode === "login" ? (
        <LoginForm onSubmit={handleLogin} />
      ) : (
        <RegisterForm onSubmit={handleRegister} />
      )}
    </LoginLayout>
  );
}

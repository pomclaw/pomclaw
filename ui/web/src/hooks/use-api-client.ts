import { useEffect } from "react";
import { useAuthStore } from "@/stores/use-auth-store";
import { setAuthConfig, setOnAuthFailure } from "@/client/auth";
import * as pomclawApi from "@/client/pomclaw";

export function useApiClient() {
  const { token, userId, senderID, tenantId, logout } = useAuthStore();

  useEffect(() => {
    setAuthConfig({ token, userId, senderID, tenantId });
  }, [token, userId, senderID, tenantId]);

  useEffect(() => {
    setOnAuthFailure(logout);
    return () => setOnAuthFailure(null);
  }, [logout]);

  return pomclawApi;
}

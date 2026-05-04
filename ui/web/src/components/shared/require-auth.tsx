import { Navigate, useLocation } from "react-router";
import { useAuthStore } from "@/stores/use-auth-store";
import { ROUTES } from "@/lib/constants";

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  const userId = useAuthStore((s) => s.userId);
  const senderID = useAuthStore((s) => s.senderID);
  const location = useLocation();

  // Not authenticated
  if ((!token && !senderID) || !userId) {
    return <Navigate to={ROUTES.LOGIN} state={{ from: location }} replace />;
  }

  // Multi-tenant logic removed - single tenant mode
  return <>{children}</>;
}

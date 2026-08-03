import { useAuthStore } from "@/stores/use-auth-store";

function SetupLoader() {
  return (
    <div className="flex h-dvh items-center justify-center">
      <div className="h-6 w-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
    </div>
  );
}

export function RequireSetup({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  const userId = useAuthStore((s) => s.userId);
  const senderID = useAuthStore((s) => s.senderID);

  // Only require valid credentials - WebSocket is only needed for chat functionality
  // and can fail gracefully with an in-page error message
  const hasCredentials = (token || senderID) && userId;

  // If user has no credentials, they should not be here (RequireAuth should have blocked)
  // But just in case, don't render anything
  if (!hasCredentials) {
    return <SetupLoader />;
  }

  // Allow access to main app - WebSocket connection is optional and chat will show errors
  return <>{children}</>;
}

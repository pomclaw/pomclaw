import { useTranslation } from "react-i18next";

export type LoginMode = "login" | "register";

interface LoginTabsProps {
  mode: LoginMode;
  onModeChange: (mode: LoginMode) => void;
}

export function LoginTabs({ mode, onModeChange }: LoginTabsProps) {
  const { t } = useTranslation("login");
  return (
    <div className="flex rounded-md border bg-muted p-1">
      <button
        type="button"
        onClick={() => onModeChange("login")}
        className={`flex-1 rounded-sm px-3 py-1.5 text-sm font-medium transition-colors ${
          mode === "login"
            ? "bg-background text-foreground shadow-sm"
            : "text-muted-foreground hover:text-foreground"
        }`}
      >
        Login
      </button>
      <button
        type="button"
        onClick={() => onModeChange("register")}
        className={`flex-1 rounded-sm px-3 py-1.5 text-sm font-medium transition-colors ${
          mode === "register"
            ? "bg-background text-foreground shadow-sm"
            : "text-muted-foreground hover:text-foreground"
        }`}
      >
        Register
      </button>
    </div>
  );
}

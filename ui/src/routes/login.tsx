import { IconLanguage, IconMoon, IconSun } from "@tabler/icons-react"
import { useNavigate } from "@tanstack/react-router"
import { createFileRoute } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useTheme } from "@/hooks/use-theme"
import { useAuth } from "@/hooks/use-auth"

function LoginPage() {
  const { t, i18n } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const { login, isAuthenticated, isLoading, error } = useAuth()
  const navigate = useNavigate()

  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [localError, setLocalError] = useState("")

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: "/" })
    }
  }, [isAuthenticated, navigate])

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLocalError("")

    if (!username.trim() || !password.trim()) {
      setLocalError(t("login.errorRequired", "Username and password are required"))
      return
    }

    const success = await login(username, password)
    if (!success) {
      setLocalError(error || t("login.errorFailed", "Login failed"))
    }
  }

  return (
    <div className="bg-background text-foreground flex min-h-dvh flex-col">
      <header className="border-border/50 flex h-14 shrink-0 items-center justify-end gap-2 border-b px-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="icon" aria-label="Language">
              <IconLanguage className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => i18n.changeLanguage("en")}>
              English
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => i18n.changeLanguage("zh")}>
              简体中文
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          variant="outline"
          size="icon"
          type="button"
          onClick={() => toggleTheme()}
          aria-label={theme === "dark" ? "Light mode" : "Dark mode"}
        >
          {theme === "dark" ? (
            <IconSun className="size-4" />
          ) : (
            <IconMoon className="size-4" />
          )}
        </Button>
      </header>

      <div className="flex flex-1 items-center justify-center p-4">
        <Card className="w-full max-w-md" size="sm">
          <CardHeader>
            <CardTitle>{t("login.title", "Login")}</CardTitle>
            <CardDescription>
              {t("login.description", "Sign in to your account")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="flex flex-col gap-4" onSubmit={onSubmit}>
              <div className="flex flex-col gap-2">
                <Label htmlFor="login-username">
                  {t("login.usernameLabel", "Username")}
                </Label>
                <Input
                  id="login-username"
                  name="username"
                  type="text"
                  autoComplete="username"
                  required
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder={t("login.usernamePlaceholder", "Enter your username")}
                  disabled={isLoading}
                />
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="login-password">
                  {t("login.passwordLabel", "Password")}
                </Label>
                <Input
                  id="login-password"
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={t("login.passwordPlaceholder", "Enter your password")}
                  disabled={isLoading}
                />
              </div>

              <Button type="submit" disabled={isLoading}>
                {isLoading ? t("labels.loading", "Loading...") : t("login.submit", "Sign In")}
              </Button>

              {error || localError ? (
                <p className="text-destructive text-sm" role="alert">
                  {error || localError}
                </p>
              ) : null}

              <div className="text-center text-sm">
                <span className="text-muted-foreground">
                  {t("login.noAccount", "Don't have an account?")}
                </span>
                {" "}
                <a
                  href="/register"
                  className="text-primary hover:underline"
                >
                  {t("login.registerLink", "Sign up")}
                </a>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export const Route = createFileRoute("/login")({
  component: LoginPage,
})

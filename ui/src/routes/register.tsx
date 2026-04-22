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

function RegisterPage() {
  const { t, i18n } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const { register, isAuthenticated, isLoading, error } = useAuth()
  const navigate = useNavigate()

  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [localError, setLocalError] = useState("")

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: "/" })
    }
  }, [isAuthenticated, navigate])

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(email)
  }

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLocalError("")

    // Client-side validation
    if (!username.trim()) {
      setLocalError(t("register.errorUsernameRequired", "Username is required"))
      return
    }

    if (username.trim().length < 3) {
      setLocalError(
        t("register.errorUsernameLength", "Username must be at least 3 characters"),
      )
      return
    }

    if (!email.trim()) {
      setLocalError(t("register.errorEmailRequired", "Email is required"))
      return
    }

    if (!validateEmail(email)) {
      setLocalError(t("register.errorEmailInvalid", "Please enter a valid email"))
      return
    }

    if (!password) {
      setLocalError(t("register.errorPasswordRequired", "Password is required"))
      return
    }

    if (password.length < 8) {
      setLocalError(
        t("register.errorPasswordLength", "Password must be at least 8 characters"),
      )
      return
    }

    if (password !== confirmPassword) {
      setLocalError(t("register.errorPasswordMatch", "Passwords do not match"))
      return
    }

    const success = await register(username, email, password)
    if (success) {
      // Redirect to login on success
      navigate({ to: "/login" })
    } else {
      setLocalError(error || t("register.errorFailed", "Registration failed"))
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
            <CardTitle>{t("register.title", "Create Account")}</CardTitle>
            <CardDescription>
              {t("register.description", "Sign up to get started")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="flex flex-col gap-4" onSubmit={onSubmit}>
              <div className="flex flex-col gap-2">
                <Label htmlFor="register-username">
                  {t("register.usernameLabel", "Username")}
                </Label>
                <Input
                  id="register-username"
                  name="username"
                  type="text"
                  autoComplete="username"
                  required
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder={t("register.usernamePlaceholder", "Choose a username")}
                  disabled={isLoading}
                />
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="register-email">
                  {t("register.emailLabel", "Email")}
                </Label>
                <Input
                  id="register-email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={t("register.emailPlaceholder", "your@email.com")}
                  disabled={isLoading}
                />
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="register-password">
                  {t("register.passwordLabel", "Password")}
                </Label>
                <Input
                  id="register-password"
                  name="password"
                  type="password"
                  autoComplete="new-password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={t("register.passwordPlaceholder", "At least 8 characters")}
                  disabled={isLoading}
                />
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="register-confirm-password">
                  {t("register.confirmPasswordLabel", "Confirm Password")}
                </Label>
                <Input
                  id="register-confirm-password"
                  name="confirmPassword"
                  type="password"
                  autoComplete="new-password"
                  required
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder={t("register.confirmPasswordPlaceholder", "Repeat your password")}
                  disabled={isLoading}
                />
              </div>

              <Button type="submit" disabled={isLoading}>
                {isLoading
                  ? t("labels.loading", "Loading...")
                  : t("register.submit", "Create Account")}
              </Button>

              {error || localError ? (
                <p className="text-destructive text-sm" role="alert">
                  {error || localError}
                </p>
              ) : null}

              <div className="text-center text-sm">
                <span className="text-muted-foreground">
                  {t("register.hasAccount", "Already have an account?")}
                </span>
                {" "}
                <a href="/login" className="text-primary hover:underline">
                  {t("register.loginLink", "Sign in")}
                </a>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export const Route = createFileRoute("/register")({
  component: RegisterPage,
})

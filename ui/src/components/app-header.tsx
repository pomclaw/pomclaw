import {
  IconBook,
  IconLanguage,
  IconLogout,
  IconMenu2,
  IconMoon,
  IconSun,
} from "@tabler/icons-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { postLauncherDashboardLogout } from "@/api/launcher-auth"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog.tsx"
import { Button } from "@/components/ui/button.tsx"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu.tsx"
import { Separator } from "@/components/ui/separator.tsx"
import { SidebarTrigger } from "@/components/ui/sidebar"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { useTheme } from "@/hooks/use-theme.ts"

export function AppHeader() {
  const { i18n, t } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const [showLogoutDialog, setShowLogoutDialog] = React.useState(false)

  const handleLogout = async () => {
    await postLauncherDashboardLogout()
    globalThis.location.assign("/launcher-login")
  }

  return (
    <header className="bg-background/95 supports-backdrop-filter:bg-background/60 border-b-border/50 sticky top-0 z-50 flex h-14 shrink-0 items-center justify-between border-b px-4 backdrop-blur">
      <div className="flex items-center gap-3">
        <img
          src="/pomclaw-logo.svg"
          alt="pomclaw"
          className="h-24 w-24"
        />
        <SidebarTrigger className="text-muted-foreground hover:bg-accent hover:text-foreground flex h-9 w-9 items-center justify-center rounded-lg sm:hidden [&>svg]:size-5">
          <IconMenu2 />
        </SidebarTrigger>
      </div>


      <AlertDialog open={showLogoutDialog} onOpenChange={setShowLogoutDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("header.logout.tooltip")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("header.logout.description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleLogout()}>
              {t("header.logout.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="text-muted-foreground flex items-center gap-1 text-sm font-medium md:gap-2">
        <Separator
          className="mx-4 my-2 hidden md:block"
          orientation="vertical"
        />

        {/* Docs Link */}
        <Button
          variant="ghost"
          size="icon"
          className="size-8"
          data-tour="docs-button"
          asChild
        >
          <a href="https://docs.picoclaw.io" target="_blank" rel="noreferrer">
            <IconBook className="size-4.5" />
          </a>
        </Button>

        {/* Language Switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="size-8">
              <IconLanguage className="size-4.5" />
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

        {/* Theme Toggle */}
        <Tooltip delayDuration={700}>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              onClick={() => setShowLogoutDialog(true)}
              aria-label={t("header.logout.tooltip")}
            >
              <IconLogout className="size-4.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t("header.logout.tooltip")}</TooltipContent>
        </Tooltip>

        <Button
          variant="ghost"
          size="icon"
          className="size-8"
          onClick={toggleTheme}
        >
          {theme === "dark" ? (
            <IconSun className="size-4.5" />
          ) : (
            <IconMoon className="size-4.5" />
          )}
        </Button>
      </div>
    </header>
  )
}

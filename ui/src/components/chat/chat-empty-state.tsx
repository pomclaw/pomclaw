import { IconRobot } from "@tabler/icons-react"
import { useTranslation } from "react-i18next"

export function ChatEmptyState() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col items-center justify-center py-20 opacity-70">
      <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-500">
        <IconRobot className="h-8 w-8" />
      </div>
      <h3 className="mb-2 text-xl font-medium">{t("chat.welcome")}</h3>
      <p className="text-muted-foreground text-center text-sm">
        {t("chat.welcomeDesc")}
      </p>
    </div>
  )
}

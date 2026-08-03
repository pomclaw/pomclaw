import { useTranslation } from "react-i18next";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";

interface ShareSectionProps {
  isShared: boolean;
  onIsSharedChange: (v: boolean) => void;
}

export function ShareSection({ isShared, onIsSharedChange }: ShareSectionProps) {
  const { t } = useTranslation("agents");

  return (
    <section className="space-y-3 rounded-lg border p-3 sm:p-4">
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-0.5">
          <Label htmlFor="share-toggle" className="text-sm font-normal">
            {t("create.isShared")}
          </Label>
          <p className="text-xs text-muted-foreground">{t("create.isSharedHint")}</p>
        </div>
        <Switch
          id="share-toggle"
          checked={isShared}
          onCheckedChange={onIsSharedChange}
        />
      </div>
    </section>
  );
}

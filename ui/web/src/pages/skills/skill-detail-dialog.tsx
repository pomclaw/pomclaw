import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Download, Loader2 } from "lucide-react";
import { MarkdownRenderer } from "@/components/shared/markdown-renderer";
import type { SkillInfo } from "@/types/skill";

interface SkillDetailDialogProps {
  skill: SkillInfo;
  onClose: () => void;
  getSkill: (id: string) => Promise<{ skill: SkillInfo; content: string }>;
}

export function SkillDetailDialog({
  skill,
  onClose,
  getSkill,
}: SkillDetailDialogProps) {
  const { t } = useTranslation("skills");
  const [readmeContent, setReadmeContent] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!skill.id) return;
    setLoading(true);
    getSkill(skill.id)
      .then((res) => setReadmeContent(res.content || null))
      .catch(() => setReadmeContent(null))
      .finally(() => setLoading(false));
  }, [skill.id, getSkill]);

  return (
    <Dialog open onOpenChange={() => onClose()}>
      <DialogContent className="max-h-[85vh] overflow-hidden flex flex-col sm:max-w-2xl md:max-w-3xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 flex-wrap">
            {skill.name}
            {skill.version != null && (
              <Badge variant="outline">v{skill.version}</Badge>
            )}
            {skill.visibility && (
              <Badge variant="secondary">{skill.visibility}</Badge>
            )}
          </DialogTitle>
          {skill.description && (
            <p className="text-sm text-muted-foreground">{skill.description}</p>
          )}
        </DialogHeader>

        <div className="flex-1 overflow-y-auto mt-2">
          {loading && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> {t("detail.loadingContent")}
            </div>
          )}
          {readmeContent && (
            <div className="rounded-md border bg-muted/30 p-4">
              <MarkdownRenderer content={readmeContent} />
            </div>
          )}
          {!readmeContent && !loading && (
            <p className="text-sm text-muted-foreground">{t("detail.noContent")}</p>
          )}
        </div>

        {skill.content_url && (
          <div className="flex items-center pt-3 border-t">
            <Button variant="outline" size="sm" className="gap-1.5" asChild>
              <a href={skill.content_url} download>
                <Download className="h-3.5 w-3.5" />
                {t("detail.downloadZip")}
              </a>
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

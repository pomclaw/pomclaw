import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Search, Wrench, AlertCircle, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/shared/empty-state";
import type { MCPToolInfo } from "@/client/pomclawComponents";

interface MCPToolsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serverName: string;
  toolPrefix: string;
  onLoadTools: () => Promise<MCPToolInfo[]>;
}

export function MCPToolsDialog({
  open,
  onOpenChange,
  serverName,
  toolPrefix,
  onLoadTools,
}: MCPToolsDialogProps) {
  const { t } = useTranslation("mcp");
  const [tools, setTools] = useState<MCPToolInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState("");

  useEffect(() => {
    if (!open) {
      setTools([]);
      setError(null);
      setFilter("");
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    onLoadTools()
      .then((result) => {
        if (!cancelled) {
          setTools(result);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : t("tools.failedLoad"));
          setLoading(false);
        }
      });

    return () => { cancelled = true; };
  }, [open, onLoadTools, t]);

  const filtered = tools.filter(
    (tool) =>
      tool.name.toLowerCase().includes(filter.toLowerCase()) ||
      (tool.description ?? "").toLowerCase().includes(filter.toLowerCase()),
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("tools.title", { name: serverName })}</DialogTitle>
        </DialogHeader>

        {toolPrefix && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>{t("tools.prefix")}:</span>
            <Badge variant="outline" className="font-mono">mcp_{toolPrefix}</Badge>
          </div>
        )}

        {loading && (
          <div className="flex flex-col items-center gap-2 py-8 text-muted-foreground">
            <Loader2 className="h-6 w-6 animate-spin" />
            <p className="text-sm">{t("tools.discovering")}</p>
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 rounded-md border border-destructive/50 bg-destructive/5 p-3 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {!loading && !error && tools.length === 0 && (
          <EmptyState
            icon={Wrench}
            title={t("tools.noToolsTitle")}
            description={t("tools.noToolsDescription")}
          />
        )}

        {!loading && !error && tools.length > 0 && (
          <div className="space-y-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                className="pl-8"
                placeholder={t("tools.filterPlaceholder")}
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
              />
            </div>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>{filtered.length} / {tools.length}</span>
            </div>
            <div className="max-h-64 space-y-1 overflow-y-auto">
              {filtered.map((tool) => (
                <div
                  key={tool.name}
                  className="rounded-md border p-2.5 text-sm"
                >
                  <div className="font-mono text-xs font-medium">{tool.name}</div>
                  {tool.description && (
                    <div className="mt-0.5 text-xs text-muted-foreground line-clamp-2">
                      {tool.description}
                    </div>
                  )}
                </div>
              ))}
              {filtered.length === 0 && (
                <p className="py-4 text-center text-sm text-muted-foreground">
                  {t("tools.noMatch")}
                </p>
              )}
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
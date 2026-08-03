import { useState, useCallback, lazy, Suspense } from "react";
import { useTranslation } from "react-i18next";
import { Zap, RefreshCw, Upload, Plug, Search, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/shared/page-header";
import { EmptyState } from "@/components/shared/empty-state";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { Pagination } from "@/components/shared/pagination";
import { ConfirmDeleteDialog } from "@/components/shared/confirm-delete-dialog";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/use-auth-store";
import { usePagination } from "@/hooks/use-pagination";
import { useSkills, type SkillInfo } from "./hooks/use-skills";
import { SkillUploadDialog } from "./skill-upload-dialog";
import { SkillEditDialog } from "./skill-edit-dialog";

const SkillGrantsDialog = lazy(() =>
  import("./skill-grants-dialog").then((m) => ({ default: m.SkillGrantsDialog })),
);

export function SkillsPage() {
  const { t } = useTranslation("skills");
  const { t: tc } = useTranslation("common");
  const userId = useAuthStore((s) => s.userId);
  const {
    skills, loading, refresh, uploadSkill, updateSkill, deleteSkill,
    listSkillGrants, grantAgent, revokeAgent,
  } = useSkills();

  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "mine" | "shared">("all");
  const [uploadOpen, setUploadOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<SkillInfo | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<SkillInfo | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [grantsDialogSkill, setGrantsDialogSkill] = useState<SkillInfo | null>(null);

  const tabbed = skills.filter((s) => {
    if (filter === "all") return true;
    if (filter === "mine") return s.created_by === userId;
    if (filter === "shared") return s.is_shared;
    return true;
  });
  const filtered = tabbed.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      (s.description ?? "").toLowerCase().includes(search.toLowerCase()),
  );

  const { pageItems, pagination, setPage, setPageSize } = usePagination(filtered, { defaultPageSize: 20 });

  const handleEdit = useCallback((skill: SkillInfo) => {
    setEditTarget(skill);
  }, []);

  const handleDelete = useCallback(async () => {
    if (!deleteTarget?.id) return;
    setDeleteLoading(true);
    try {
      await deleteSkill(deleteTarget.id);
      setDeleteTarget(null);
    } finally {
      setDeleteLoading(false);
    }
  }, [deleteTarget, deleteSkill]);

  return (
    <div className="flex h-full flex-col gap-4 p-4">
      <PageHeader
        title={t("title")}
        description={t("description")}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setUploadOpen(true)}>
              <Upload className="mr-1.5 h-4 w-4" />
              {t("upload.button")}
            </Button>
            <Button variant="outline" size="sm" onClick={refresh}>
              <RefreshCw className="mr-1.5 h-4 w-4" />
              {tc("refresh")}
            </Button>
          </div>
        }
      />

      {/* Filter tabs */}
      <div className="flex gap-1 border-b">
        {(["all", "mine", "shared"] as const).map((f) => (
          <button
            key={f}
            type="button"
            className={cn(
              "px-3 py-1.5 text-sm font-medium border-b-2 -mb-px transition-colors",
              filter === f
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
            onClick={() => { setFilter(f); setPage(1); }}
          >
            {t(`filter.${f}`)}
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="relative w-full max-w-sm">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          className="pl-8"
          placeholder={t("searchPlaceholder")}
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPage(1);
          }}
        />
      </div>

      {/* Loading */}
      {loading && (
        <TableSkeleton rows={5} />
      )}

      {/* Empty state */}
      {!loading && skills.length === 0 && (
        <EmptyState
          icon={Zap}
          title={t("emptyTitle")}
          description={t("emptyDescription")}
          action={<Button onClick={() => setUploadOpen(true)}><Upload className="mr-1.5 h-4 w-4" />{t("upload.button")}</Button>}
        />
      )}

      {/* Filtered empty */}
      {!loading && skills.length > 0 && filtered.length === 0 && (
        <EmptyState
          icon={Search}
          title={t("noMatchTitle")}
          description={t("noMatchDescription")}
        />
      )}

      {/* Table */}
      {!loading && filtered.length > 0 && (
        <div className="min-w-0 overflow-x-auto rounded-md border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.name")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.description")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.agents")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.enabled")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.createdBy")}</th>
                <th className="px-3 py-2.5 text-right font-medium text-muted-foreground">{t("columns.actions")}</th>
              </tr>
            </thead>
            <tbody>
              {pageItems.map((skill) => (
                <tr key={skill.id || skill.name} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{skill.name}</span>
                      {skill.is_shared && (
                        <Badge variant="secondary" className="h-5 px-1.5 text-2xs">
                          {tc("shared")}
                        </Badge>
                      )}
                      {skill.version && (
                        <span className="text-xs text-muted-foreground">v{skill.version}</span>
                      )}
                    </div>
                    {skill.missing_deps && skill.missing_deps.length > 0 && (
                      <div className="text-xs text-amber-600 dark:text-amber-400 mt-.5">
                        {t("deps.missing", { deps: skill.missing_deps.slice(0, 2).join(", ") })}
                      </div>
                    )}
                  </td>
                  <td className="px-3 py-2.5 text-muted-foreground max-w-xs truncate">
                    {skill.description || "-"}
                  </td>
                  <td className="px-3 py-2.5">
                    <span className="text-muted-foreground">{skill.agent_count ?? 0}</span>
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge variant={skill.enabled !== false ? "default" : "secondary"} className="text-xs">
                      {skill.enabled !== false ? "Yes" : "No"}
                    </Badge>
                  </td>
                  <td className="px-3 py-2.5 text-muted-foreground">
                    {skill.created_by_name || skill.created_by || "-"}
                  </td>
                  <td className="px-3 py-2.5">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setGrantsDialogSkill(skill)}
                        title={t("grants.title", { name: "" })}
                      >
                        <Plug className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => handleEdit(skill)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setDeleteTarget(skill)}
                      >
                        <Trash2 className="h-3.5 w-3.5 text-destructive" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <Pagination
            page={pagination.page}
            pageSize={pagination.pageSize}
            total={pagination.total}
            totalPages={pagination.totalPages}
            onPageChange={setPage}
            onPageSizeChange={setPageSize}
          />
        </div>
      )}

      {/* Dialogs */}
      <SkillUploadDialog open={uploadOpen} onOpenChange={setUploadOpen} onUpload={(f) => uploadSkill(f)} />

      {editTarget && (
        <SkillEditDialog
          skill={editTarget}
          onClose={() => setEditTarget(null)}
          onSave={async (id, updates) => { await updateSkill(id, updates); setEditTarget(null); }}
        />
      )}

      {grantsDialogSkill && (
        <Suspense fallback={null}>
          <SkillGrantsDialog
            open={!!grantsDialogSkill}
            onOpenChange={(open) => { if (!open) setGrantsDialogSkill(null); }}
            skillId={grantsDialogSkill.id!}
            skillName={grantsDialogSkill.name}
            onLoadGrants={listSkillGrants}
            onGrantAgent={grantAgent}
            onRevokeAgent={revokeAgent}
          />
        </Suspense>
      )}

      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
        title={t("delete.title")}
        description={t("delete.description", { name: deleteTarget?.name ?? "" })}
        confirmValue={deleteTarget?.name ?? ""}
        confirmLabel={t("delete.confirmLabel")}
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  );
}
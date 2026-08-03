import { useState, useCallback, lazy, Suspense } from "react";
import { useTranslation } from "react-i18next";
import { Plus, RefreshCw, Plug, Search, Wrench, Pencil, Trash2 } from "lucide-react";
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
import { useMCP } from "./hooks/use-mcp";
import { MCPFormDialog } from "./mcp-form-dialog";
import { MCPToolsDialog } from "./mcp-tools-dialog";
import type { MCPServer } from "@/client/pomclawComponents";

const MCPGrantsDialog = lazy(() =>
  import("./mcp-grants-dialog").then((m) => ({ default: m.MCPGrantsDialog })),
);

export function MCPPage() {
  const { t } = useTranslation("mcp");
  const { t: tc } = useTranslation("common");
  const userId = useAuthStore((s) => s.userId);
  const {
    servers,
    loading,
    refresh,
    createServer,
    updateServer,
    deleteServer,
    testConnection,
    reconnectServer,
    listServerTools,
    listServerGrants,
    grantAgent,
    revokeAgent,
  } = useMCP();

  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "mine" | "shared">("all");
  const [formOpen, setFormOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<MCPServer | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<MCPServer | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [toolsDialogServer, setToolsDialogServer] = useState<MCPServer | null>(null);
  const [grantsDialogServer, setGrantsDialogServer] = useState<MCPServer | null>(null);

  const tabbed = servers.filter((s) => {
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

  const handleAdd = useCallback(() => {
    setEditingServer(null);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((server: MCPServer) => {
    setEditingServer(server);
    setFormOpen(true);
  }, []);

  const handleFormSave = useCallback(
    async (data: Record<string, unknown>) => {
      const transport = data.transport as string;
      const isStdio = transport === "stdio";

      const command = isStdio ? (data.command as string) : undefined;
      const args = isStdio && data.args
        ? JSON.stringify((data.args as string).split(/\s+/).filter(Boolean))
        : undefined;
      const url = !isStdio ? (data.url as string) : undefined;
      const headers = !isStdio && data.headers && Object.keys(data.headers as Record<string, string>).length > 0
        ? JSON.stringify(data.headers)
        : undefined;
      const env = data.env && Object.keys(data.env as Record<string, string>).length > 0
        ? JSON.stringify(data.env)
        : undefined;
      const settings = data.requireUserCreds
        ? JSON.stringify({ require_user_credentials: true })
        : undefined;

      const payload = {
        name: data.name as string,
        description: (data.description as string) || undefined,
        transport,
        command,
        args,
        url,
        headers,
        env,
        tool_prefix: (data.toolPrefix as string) || undefined,
        timeout_sec: data.timeout as number,
        enabled: data.enabled as boolean,
        settings,
      };

      if (editingServer) {
        await updateServer(editingServer.id, payload);
      } else {
        await createServer(payload);
      }
    },
    [editingServer, createServer, updateServer],
  );

  const handleTest = useCallback(
    async (data: Record<string, unknown>) => {
      const transport = data.transport as string;
      const isStdio = transport === "stdio";
      const command = isStdio ? (data.command as string) : undefined;
      const args = isStdio && data.args
        ? JSON.stringify((data.args as string).split(/\s+/).filter(Boolean))
        : undefined;
      const url = !isStdio ? (data.url as string) : undefined;
      const headers = !isStdio && data.headers && Object.keys(data.headers as Record<string, string>).length > 0
        ? JSON.stringify(data.headers)
        : undefined;
      const env = data.env && Object.keys(data.env as Record<string, string>).length > 0
        ? JSON.stringify(data.env)
        : undefined;

      return testConnection({
        name: data.name as string,
        transport,
        command,
        args,
        url,
        headers,
        env,
        timeout_sec: data.timeout as number,
      });
    },
    [testConnection],
  );

  const handleDelete = useCallback(async () => {
    if (!deleteTarget) return;
    setDeleteLoading(true);
    try {
      await deleteServer(deleteTarget.id);
      setDeleteTarget(null);
    } finally {
      setDeleteLoading(false);
    }
  }, [deleteTarget, deleteServer]);

  const handleReconnect = useCallback(
    async (id: string) => {
      await reconnectServer(id);
    },
    [reconnectServer],
  );

  return (
    <div className="flex h-full flex-col gap-4 p-4">
      <PageHeader
        title={t("title")}
        description={t("description")}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={refresh}>
              <RefreshCw className="mr-1.5 h-4 w-4" />
              Refresh
            </Button>
            <Button size="sm" onClick={handleAdd}>
              <Plus className="mr-1.5 h-4 w-4" />
              {t("addServer")}
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
      {!loading && servers.length === 0 && (
        <EmptyState
          icon={Plug}
          title={t("emptyTitle")}
          description={t("emptyDescription")}
          action={<Button onClick={handleAdd}><Plus className="mr-1.5 h-4 w-4" />{t("addServer")}</Button>}
        />
      )}

      {/* Filtered empty */}
      {!loading && servers.length > 0 && filtered.length === 0 && (
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
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.transport")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.tools")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.agents")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.enabled")}</th>
                <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.createdBy")}</th>
                <th className="px-3 py-2.5 text-right font-medium text-muted-foreground">{t("columns.actions")}</th>
              </tr>
            </thead>
            <tbody>
              {pageItems.map((server) => (
                <tr key={server.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{server.name}</span>
                      {server.is_shared && (
                        <Badge variant="secondary" className="h-5 px-1.5 text-2xs">
                          {tc("shared")}
                        </Badge>
                      )}
                    </div>
                    {server.description && (
                      <div className="text-xs text-muted-foreground">{server.description}</div>
                    )}
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge variant="outline" className="font-mono text-xs">
                      {server.transport}
                    </Badge>
                  </td>
                  <td className="px-3 py-2.5 text-muted-foreground">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setToolsDialogServer(server)}
                    >
                      <Wrench className="mr-1 h-3.5 w-3.5" />
                      {t("viewTools")}
                    </Button>
                  </td>
                  <td className="px-3 py-2.5">
                    <span className="text-muted-foreground">{server.agent_count ?? 0}</span>
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge variant={server.enabled ? "default" : "secondary"} className="text-xs">
                      {server.enabled ? "Yes" : "No"}
                    </Badge>
                  </td>
                  <td className="px-3 py-2.5 text-muted-foreground">
                    {server.created_by_name || server.created_by || "-"}
                  </td>
                  <td className="px-3 py-2.5">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => handleReconnect(server.id)}
                        title={t("reconnect")}
                      >
                        <RefreshCw className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setGrantsDialogServer(server)}
                        title={t("manageGrants")}
                      >
                        <Plug className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => handleEdit(server)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setDeleteTarget(server)}
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
      <MCPFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        server={editingServer}
        onSave={handleFormSave}
        onTest={handleTest}
      />

      <MCPToolsDialog
        open={!!toolsDialogServer}
        onOpenChange={(open) => { if (!open) setToolsDialogServer(null); }}
        serverName={toolsDialogServer?.name || ""}
        toolPrefix={toolsDialogServer?.tool_prefix || ""}
        onLoadTools={async () => {
          if (!toolsDialogServer) return [];
          return listServerTools(toolsDialogServer.id);
        }}
      />

      {grantsDialogServer && (
        <Suspense fallback={null}>
          <MCPGrantsDialog
            open={!!grantsDialogServer}
            onOpenChange={(open) => { if (!open) setGrantsDialogServer(null); }}
            serverId={grantsDialogServer.id}
            serverName={grantsDialogServer.name}
            onLoadGrants={listServerGrants}
            onGrantAgent={grantAgent}
            onRevokeAgent={revokeAgent}
            onLoadTools={listServerTools}
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
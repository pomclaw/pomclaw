import { useEffect, useMemo, useState, lazy, Suspense } from "react";
import { useParams, useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import { Cpu, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EmptyState } from "@/components/shared/empty-state";
import { SearchInput } from "@/components/shared/search-input";
import { Pagination } from "@/components/shared/pagination";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { ConfirmDeleteDialog } from "@/components/shared/confirm-delete-dialog";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/use-auth-store";
import { useProviders, type Provider } from "./hooks/use-providers";
import { useChatGPTOAuthProviderQuotas } from "./hooks/use-chatgpt-oauth-provider-quotas";
import { useChatGPTOAuthProviderStatuses } from "./hooks/use-chatgpt-oauth-provider-statuses";
import { ProviderListRow } from "./provider-list-row";

const ProviderFormDialog = lazy(() =>
  import("./provider-form-dialog").then((m) => ({ default: m.ProviderFormDialog }))
);
const PoolSetupWizardDialog = lazy(() =>
  import("./pool-setup-wizard-dialog").then((m) => ({ default: m.PoolSetupWizardDialog }))
);
import {
  getChatGPTOAuthPoolOwnership,
  sortProvidersForPoolHierarchy,
} from "./provider-utils";
import { useDeferredLoading } from "@/hooks/use-deferred-loading";
import { usePagination } from "@/hooks/use-pagination";
import { ProviderDetailPage } from "./provider-detail/provider-detail-page";

export function ProvidersPage() {
  const { id: detailId } = useParams<{ id: string }>();
  const navigate = useNavigate();

  if (detailId) {
    return (
      <ProviderDetailPage
        providerId={detailId}
        onBack={() => navigate("/providers")}
      />
    );
  }

  return <ProviderListView />;
}


function ProviderListView() {
  const { t } = useTranslation("providers");
  const navigate = useNavigate();
  const userId = useAuthStore((s) => s.userId);

  const {
    providers, loading, refresh,
    createProvider, updateProvider, deleteProvider,
  } = useProviders();
  const showSkeleton = useDeferredLoading(loading && providers.length === 0);
  const { statuses } = useChatGPTOAuthProviderStatuses(providers);

  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "mine" | "shared">("all");
  const [formOpen, setFormOpen] = useState(false);
  const [wizardOpen, setWizardOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Provider | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const providerByName = useMemo(
    () => new Map(providers.map((provider) => [provider.name, provider])),
    [providers],
  );
  const poolOwnership = useMemo(
    () => getChatGPTOAuthPoolOwnership(providers),
    [providers],
  );
  const selectablePoolOwnership = useMemo(
    () => getChatGPTOAuthPoolOwnership(providers, { enabledOnly: true }),
    [providers],
  );
  const oauthAvailabilityByName = useMemo(
    () => new Map(statuses.map((status) => [status.provider.name, status.availability])),
    [statuses],
  );
  // Unpooled = chatgpt_oauth providers that are neither pool owners nor members
  const unpooledProviders = useMemo(
    () => providers.filter(
      (p) =>
        p.provider_type === "chatgpt_oauth" &&
        p.enabled &&
        !selectablePoolOwnership.membersByOwner.has(p.name) &&
        !selectablePoolOwnership.ownerByMember.has(p.name),
    ),
    [providers, selectablePoolOwnership],
  );

  const tabbed = useMemo(() => providers.filter((p) => {
    if (filter === "all") return true;
    if (filter === "mine") return p.created_by === userId;
    if (filter === "shared") return p.is_shared;
    return true;
  }), [providers, filter, userId]);
  const filtered = useMemo(() => tabbed.filter(
    (provider) =>
      provider.name.toLowerCase().includes(search.toLowerCase()) ||
      (provider.description ?? "").toLowerCase().includes(search.toLowerCase()),
  ), [tabbed, search]);
  const orderedProviders = useMemo(
    () => sortProvidersForPoolHierarchy(filtered, poolOwnership),
    [filtered, poolOwnership],
  );
  const { pageItems, pagination, setPage, setPageSize, resetPage } = usePagination(orderedProviders);
  const visibleQuotaProviderNames = useMemo(
    () =>
      pageItems
        .filter(
          (provider) =>
            provider.provider_type === "chatgpt_oauth" &&
            oauthAvailabilityByName.get(provider.name) === "ready",
        )
        .map((provider) => provider.name),
    [oauthAvailabilityByName, pageItems],
  );
  const {
    quotaByName,
    isLoading: quotasLoading,
    isFetching: quotasFetching,
  } = useChatGPTOAuthProviderQuotas(visibleQuotaProviderNames, visibleQuotaProviderNames.length > 0);

  useEffect(() => { resetPage(); }, [search, filter, resetPage]);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleteLoading(true);
    try {
      await deleteProvider(deleteTarget.id);
      setDeleteTarget(null);
    } finally {
      setDeleteLoading(false);
    }
  };

  return (
    <div className="p-4 sm:p-6 pb-10">
      <PageHeader
        title={t("title")}
        description={t("description")}
        actions={
          <Button onClick={() => setFormOpen(true)} className="gap-1">
            <Plus className="h-4 w-4" /> {t("addProvider")}
          </Button>
        }
      />

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder={t("searchPlaceholder")}
          className="max-w-sm"
        />
      </div>

      {/* Filter tabs */}
      <div className="mt-4 flex gap-1 border-b">
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

      <div className="mt-6">
        {showSkeleton ? (
          <TableSkeleton />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Cpu}
            title={search ? t("noMatchTitle") : t("emptyTitle")}
            description={search ? t("noMatchDescription") : t("emptyDescription")}
          />
        ) : (
          <>
            <div className="min-w-0 overflow-x-auto rounded-md border">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.name")}</th>
                    <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.type")}</th>
                    <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.apiKey")}</th>
                    <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.status")}</th>
                    <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">{t("columns.createdBy")}</th>
                    <th className="px-3 py-2.5 text-right font-medium text-muted-foreground">{t("columns.actions")}</th>
                  </tr>
                </thead>
                <tbody>
                  {pageItems.map((p) => (
                    <ProviderListRow
                      key={p.id}
                      provider={p}
                      oauthPool={p.provider_type === "chatgpt_oauth" ? {
                        availability: oauthAvailabilityByName.get(p.name) ?? (p.enabled ? "needs_sign_in" : "disabled"),
                        role: poolOwnership.ownerByMember.has(p.name)
                          ? "member"
                          : poolOwnership.membersByOwner.has(p.name)
                            ? "owner"
                            : "standalone",
                        managedByLabel: (() => {
                          const ownerName = poolOwnership.ownerByMember.get(p.name);
                          if (!ownerName) return undefined;
                          const owner = providerByName.get(ownerName);
                          return owner?.name || ownerName;
                        })(),
                        memberCount: poolOwnership.membersByOwner.get(p.name)?.length ?? 0,
                        strategy: poolOwnership.strategyByOwner.get(p.name) ?? "priority_order",
                        quota: quotaByName.get(p.name),
                        quotaLoading: oauthAvailabilityByName.get(p.name) === "ready"
                          ? quotasLoading || quotasFetching
                          : false,
                      } : undefined}
                      showPoolHint={
                        p.provider_type === "chatgpt_oauth" &&
                        p.enabled &&
                        !selectablePoolOwnership.ownerByMember.has(p.name) &&
                        !selectablePoolOwnership.membersByOwner.has(p.name) &&
                        unpooledProviders.length >= 2
                      }
                      onClick={() => navigate(`/providers/${p.id}`)}
                      onDelete={() => setDeleteTarget(p)}
                      onPoolSetup={() => setWizardOpen(true)}
                    />
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
          </>
        )}
      </div>

      <Suspense fallback={null}>
        <ProviderFormDialog
          open={formOpen}
          onOpenChange={setFormOpen}
          onSubmit={async (data) => {
            await createProvider(data);
            refresh();
          }}
          existingProviders={providers}
        />
      </Suspense>

      <Suspense fallback={null}>
        <PoolSetupWizardDialog
          open={wizardOpen}
          onOpenChange={setWizardOpen}
          providers={providers}
          unpooledProviders={unpooledProviders}
          onSave={async (ownerId, data) => {
            await updateProvider(ownerId, data);
            refresh();
          }}
        />
      </Suspense>

      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title={t("delete.title")}
        description={t("delete.description", { name: deleteTarget?.name })}
        confirmValue={deleteTarget?.name || ""}
        confirmLabel={t("delete.confirmLabel")}
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  );
}

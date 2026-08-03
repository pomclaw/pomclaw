import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Loader2, CheckCircle, AlertCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { McpConnectionFields } from "./mcp-connection-fields";
import { McpSettingsFields } from "./mcp-settings-fields";
import { mcpFormSchema, type MCPFormData } from "./schemas/mcp.schema";
import type { MCPServer } from "@/client/pomclawComponents";

interface MCPFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  server?: MCPServer | null;
  onSave: (data: MCPFormData) => Promise<void>;
  onTest: (data: MCPFormData) => Promise<{ success: boolean; tool_count?: number; error?: string }>;
}

function serverToFormData(server: MCPServer): MCPFormData {
  let args = "";
  if (server.args) {
    try {
      args = (JSON.parse(server.args) as string[]).join(" ");
    } catch { /* ignore */ }
  }
  let headers: Record<string, string> = {};
  if (server.headers) {
    try { headers = JSON.parse(server.headers); } catch { /* ignore */ }
  }
  let env: Record<string, string> = {};
  if (server.env) {
    try { env = JSON.parse(server.env); } catch { /* ignore */ }
  }
  let requireUserCreds = false;
  if (server.settings) {
    try {
      const s = JSON.parse(server.settings);
      requireUserCreds = !!s.require_user_credentials;
    } catch { /* ignore */ }
  }

  return {
    name: server.name,
    description: server.description || "",
    transport: server.transport as "stdio" | "sse" | "streamable-http",
    command: server.command || "",
    args,
    url: server.url || "",
    headers,
    env,
    toolPrefix: server.tool_prefix || "",
    timeout: server.timeout_sec || 60,
    enabled: server.enabled,
    requireUserCreds,
  };
}

export function MCPFormDialog({ open, onOpenChange, server, onSave, onTest }: MCPFormDialogProps) {
  const { t } = useTranslation("mcp");
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; tool_count?: number; error?: string } | null>(null);
  const [saving, setSaving] = useState(false);

  const form = useForm<MCPFormData>({
    resolver: zodResolver(mcpFormSchema),
    defaultValues: {
      name: "",
      description: "",
      transport: "stdio",
      command: "",
      args: "",
      url: "",
      headers: {},
      env: {},
      toolPrefix: "",
      timeout: 60,
      enabled: true,
      requireUserCreds: false,
    },
  });

  useEffect(() => {
    if (open) {
      if (server) {
        form.reset(serverToFormData(server));
      } else {
        form.reset({
          name: "",
          description: "",
          transport: "stdio",
          command: "",
          args: "",
          url: "",
          headers: {},
          env: {},
          toolPrefix: "",
          timeout: 60,
          enabled: true,
          requireUserCreds: false,
        });
      }
      setTestResult(null);
    }
  }, [open, server, form]);

  const handleTest = useCallback(async () => {
    const data = form.getValues();
    setTesting(true);
    setTestResult(null);
    try {
      const result = await onTest(data);
      setTestResult(result);
    } catch (err) {
      setTestResult({ success: false, error: err instanceof Error ? err.message : "Connection failed" });
    } finally {
      setTesting(false);
    }
  }, [form, onTest]);

  const handleSubmit = useCallback(
    async (data: MCPFormData) => {
      setSaving(true);
      try {
        await onSave(data);
        onOpenChange(false);
      } finally {
        setSaving(false);
      }
    },
    [onSave, onOpenChange],
  );

  const isEditing = !!server;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{isEditing ? t("form.editTitle") : t("form.createTitle")}</DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[70vh]">
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6 px-1">
            <McpConnectionFields form={form} />
            <McpSettingsFields form={form} />

            {/* Test Connection */}
            <div className="space-y-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleTest}
                disabled={testing}
              >
                {testing && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
                {testing ? t("form.testing") : t("form.testConnection")}
              </Button>
              {testResult && (
                <div className="flex items-center gap-2 text-sm">
                  {testResult.success ? (
                    <>
                      <CheckCircle className="h-4 w-4 text-green-500" />
                      <span className="text-green-600">
                        {t("form.toolsFound", { count: testResult.tool_count ?? 0 })}
                      </span>
                    </>
                  ) : (
                    <>
                      <AlertCircle className="h-4 w-4 text-destructive" />
                      <span className="text-destructive">{testResult.error}</span>
                    </>
                  )}
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-2 border-t pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                {t("form.cancel")}
              </Button>
              <Button type="submit" disabled={saving}>
                {saving && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
                {isEditing ? t("form.update") : t("form.create")}
              </Button>
            </div>
          </form>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
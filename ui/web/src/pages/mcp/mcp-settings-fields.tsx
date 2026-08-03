import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { KeyValueEditor } from "@/components/shared/key-value-editor";
import type { UseFormReturn } from "react-hook-form";
import type { MCPFormData } from "./schemas/mcp.schema";

function isSensitiveEnv(key: string): boolean {
  return /^(api_key|api_key|secret|password|token|credential|auth|private_key)$/i.test(key);
}

interface McpSettingsFieldsProps {
  form: UseFormReturn<MCPFormData>;
}

export function McpSettingsFields({ form }: McpSettingsFieldsProps) {
  const { t } = useTranslation("mcp");
  const { register, setValue, watch } = form;
  const toolPrefix = watch("toolPrefix");

  return (
    <div className="space-y-4">
      {/* Environment Variables */}
      <div className="space-y-2">
        <Label>{t("form.env")}</Label>
        <KeyValueEditor
          value={watch("env")}
          onChange={(v) => setValue("env", v)}
          keyPlaceholder="Key"
          valuePlaceholder="Value"
          addLabel="Add Variable"
          maskValue={isSensitiveEnv}
        />
      </div>

      {/* Tool Prefix */}
      <div className="space-y-2">
        <Label htmlFor="toolPrefix">{t("form.toolPrefix")}</Label>
        <div className="relative">
          <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm text-muted-foreground">
            mcp_
          </span>
          <Input
            id="toolPrefix"
            {...register("toolPrefix")}
            className="pl-10 font-mono"
            placeholder="my-server"
          />
        </div>
        {toolPrefix && (
          <p className="text-xs text-muted-foreground">
            mcp_{toolPrefix}.*
          </p>
        )}
      </div>

      {/* Timeout */}
      <div className="space-y-2">
        <Label htmlFor="timeout">{t("form.timeout")}</Label>
        <Input
          id="timeout"
          type="number"
          min={1}
          {...register("timeout", { valueAsNumber: true })}
        />
      </div>

      {/* Enabled */}
      <div className="flex items-center justify-between">
        <Label htmlFor="enabled">{t("form.enabled")}</Label>
        <Switch
          id="enabled"
          checked={watch("enabled")}
          onCheckedChange={(v) => setValue("enabled", v)}
        />
      </div>

      {/* Require User Credentials */}
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <Label htmlFor="requireUserCreds">{t("form.requireUserCredentials")}</Label>
          <p className="text-xs text-muted-foreground">{t("form.requireUserCredentialsHint")}</p>
        </div>
        <Switch
          id="requireUserCreds"
          checked={watch("requireUserCreds")}
          onCheckedChange={(v) => setValue("requireUserCreds", v)}
        />
      </div>
    </div>
  );
}
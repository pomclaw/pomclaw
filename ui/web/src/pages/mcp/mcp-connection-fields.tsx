import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ButtonGroup, ButtonGroupItem } from "@/components/ui/button-group";
import { KeyValueEditor } from "@/components/shared/key-value-editor";
import type { UseFormReturn } from "react-hook-form";
import type { MCPFormData } from "./schemas/mcp.schema";

function isSensitiveHeader(key: string): boolean {
  return /^(authorization|x-api-key|api-key|bearer|token|secret|password|credential)$/i.test(key);
}

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9-_\s]/g, "")
    .replace(/\s+/g, "-");
}

interface McpConnectionFieldsProps {
  form: UseFormReturn<MCPFormData>;
}

export function McpConnectionFields({ form }: McpConnectionFieldsProps) {
  const { t } = useTranslation("mcp");
  const { register, setValue, watch } = form;
  const transport = watch("transport");

  const handleNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const slugified = slugify(e.target.value);
      setValue("name", slugified, { shouldValidate: true });
    },
    [setValue],
  );

  return (
    <div className="space-y-4">
      {/* Name */}
      <div className="space-y-2">
        <Label htmlFor="name">{t("form.name")}</Label>
        <Input
          id="name"
          {...register("name", { onChange: handleNameChange })}
          placeholder="my-mcp-server"
        />
        <p className="text-xs text-muted-foreground">{t("form.nameHint")}</p>
      </div>

      {/* Description */}
      <div className="space-y-2">
        <Label htmlFor="description">{t("form.description")}</Label>
        <Textarea
          id="description"
          {...register("description")}
          placeholder="My MCP Server"
          size="sm"
        />
      </div>

      {/* Transport */}
      <div className="space-y-2">
        <Label>{t("form.transport")}</Label>
        <ButtonGroup>
          <ButtonGroupItem
            active={transport === "stdio"}
            onClick={() => setValue("transport", "stdio")}
          >
            stdio
          </ButtonGroupItem>
          <ButtonGroupItem
            active={transport === "sse"}
            onClick={() => setValue("transport", "sse")}
          >
            SSE
          </ButtonGroupItem>
          <ButtonGroupItem
            active={transport === "streamable-http"}
            onClick={() => setValue("transport", "streamable-http")}
          >
            Streamable HTTP
          </ButtonGroupItem>
        </ButtonGroup>
      </div>

      {/* stdio: Command + Args */}
      {transport === "stdio" && (
        <>
          <div className="space-y-2">
            <Label htmlFor="command">{t("form.command")}</Label>
            <Input
              id="command"
              {...register("command")}
              placeholder="npx"
              className="font-mono"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="args">{t("form.args")}</Label>
            <Input
              id="args"
              {...register("args")}
              placeholder="-y @modelcontextprotocol/server-filesystem /path"
              className="font-mono"
            />
          </div>
        </>
      )}

      {/* SSE / HTTP: URL + Headers */}
      {(transport === "sse" || transport === "streamable-http") && (
        <>
          <div className="space-y-2">
            <Label htmlFor="url">{t("form.url")}</Label>
            <Input
              id="url"
              {...register("url")}
              placeholder="http://localhost:3000/sse"
              className="font-mono"
            />
          </div>
          <div className="space-y-2">
            <Label>{t("form.headers")}</Label>
            <KeyValueEditor
              value={watch("headers")}
              onChange={(v) => setValue("headers", v)}
              keyPlaceholder="Header"
              valuePlaceholder="Value"
              addLabel="Add Header"
              maskValue={isSensitiveHeader}
            />
          </div>
        </>
      )}
    </div>
  );
}
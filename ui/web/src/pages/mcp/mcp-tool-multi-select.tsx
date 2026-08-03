import { useState, useRef, useLayoutEffect, useCallback } from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { usePortalDropdownClose } from "@/hooks/use-portal-dropdown-close";

interface ToolOption {
  name: string;
  description?: string;
}

interface ToolMultiSelectProps {
  options: ToolOption[];
  value: string[];
  onChange: (value: string[]) => void;
  placeholder?: string;
}

export function ToolMultiSelect({ options, value, onChange, placeholder }: ToolMultiSelectProps) {
  const [inputValue, setInputValue] = useState("");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [dropdownStyle, setDropdownStyle] = useState<React.CSSProperties>({});

  const filtered = options.filter(
    (o) =>
      !value.includes(o.name) &&
      (o.name.toLowerCase().includes(inputValue.toLowerCase()) ||
        (o.description ?? "").toLowerCase().includes(inputValue.toLowerCase())),
  );

  usePortalDropdownClose({
    open,
    onClose: () => setOpen(false),
    ignore: [containerRef, dropdownRef],
  });

  useLayoutEffect(() => {
    if (open && containerRef.current) {
      const rect = containerRef.current.getBoundingClientRect();
      setDropdownStyle({
        position: "fixed",
        left: rect.left,
        top: rect.bottom + 4,
        width: rect.width,
        zIndex: 9999,
      });
    }
  }, [open]);

  const addValue = useCallback(
    (name: string) => {
      if (!value.includes(name)) {
        onChange([...value, name]);
      }
      setInputValue("");
    },
    [value, onChange],
  );

  const removeValue = useCallback(
    (name: string) => {
      onChange(value.filter((v) => v !== name));
    },
    [value, onChange],
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      const trimmed = inputValue.trim();
      if (trimmed) {
        addValue(trimmed);
      }
    }
    if (e.key === "Backspace" && !inputValue && value.length > 0) {
      removeValue(value[value.length - 1]!);
    }
  };

  const portalContainer = containerRef.current?.closest('[role="dialog"]') || document.body;

  return (
    <div ref={containerRef} className="relative">
      <div className="flex min-h-10 flex-wrap items-center gap-1 rounded-md border border-input bg-background px-3 py-1.5 text-sm focus-within:ring-1 focus-within:ring-ring">
        {value.map((v) => (
          <Badge key={v} variant="secondary" className="gap-1 pr-1">
            {v}
            <button
              type="button"
              className="ml-0.5 rounded-full p-0.5 hover:bg-muted"
              onClick={() => removeValue(v)}
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        ))}
        <input
          className="min-w-[60px] flex-1 border-0 bg-transparent p-0 text-sm outline-none placeholder:text-muted-foreground"
          value={inputValue}
          onChange={(e) => {
            setInputValue(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
        />
      </div>
      {open && filtered.length > 0 &&
        createPortal(
          <div
            ref={dropdownRef}
            style={dropdownStyle}
            className="max-h-48 overflow-auto rounded-md border bg-popover p-1 shadow-md"
          >
            {filtered.map((o) => (
              <button
                key={o.name}
                type="button"
                className="flex w-full flex-col gap-0.5 rounded-sm px-2 py-1.5 text-left text-sm hover:bg-accent"
                onClick={() => addValue(o.name)}
              >
                <span className="font-mono text-xs">{o.name}</span>
                {o.description && (
                  <span className="text-xs text-muted-foreground line-clamp-1">{o.description}</span>
                )}
              </button>
            ))}
          </div>,
          portalContainer,
        )}
    </div>
  );
}
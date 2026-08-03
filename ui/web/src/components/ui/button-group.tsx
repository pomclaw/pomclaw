import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface ButtonGroupProps {
  children: React.ReactNode;
  className?: string;
}

export function ButtonGroup({ children, className }: ButtonGroupProps) {
  return (
    <div className={cn("flex items-center gap-0 rounded-md border bg-muted/20 p-.5", className)}>
      {children}
    </div>
  );
}

interface ButtonGroupItemProps {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}

export function ButtonGroupItem({ active, onClick, children }: ButtonGroupItemProps) {
  return (
    <Button
      type="button"
      variant={active ? "default" : "ghost"}
      size="sm"
      onClick={onClick}
      className={cn(
        "rounded-sm px-3 text-xs",
        active ? "shadow-sm" : "text-muted-foreground hover:text-foreground",
      )}
    >
      {children}
    </Button>
  );
}
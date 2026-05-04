import { Info } from "lucide-react";
import { Label } from "@/components/ui/label";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Label with an (i) tooltip icon for field descriptions. */
export function InfoLabel({ children, tip }: { children: React.ReactNode; tip: string }) {
  return (
    <div className="flex items-center gap-1.5">
      <Label>{children}</Label>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 shrink-0 cursor-help text-muted-foreground" />
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-64">
            {tip}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}

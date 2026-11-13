import { ChevronLeft, ChevronRight } from "lucide-react";
import { useId } from "react";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

type PaginationProps = {
  total: number;
  limit: number;
  offset: number;
  onOffsetChange: (offset: number) => void;
  onLimitChange: (limit: number) => void;
  pageSizeOptions?: readonly number[];
  className?: string;
};

export function Pagination({
  total,
  limit,
  offset,
  onOffsetChange,
  onLimitChange,
  pageSizeOptions = [10, 20, 50, 100],
  className,
}: PaginationProps) {
  const hasRows = total > 0;
  const start = hasRows ? offset + 1 : 0;
  const end = hasRows ? Math.min(offset + limit, total) : 0;
  const canGoBack = offset > 0;
  const canGoForward = offset + limit < total;
  const pageSizeId = useId();

  return (
    <div
      className={cn(
        "flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between",
        className,
      )}
    >
      <p className="text-sm text-muted-foreground">
        {start}–{end} of {total}
      </p>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground" id={pageSizeId}>
            Rows per page
          </span>
          <Select value={String(limit)} onValueChange={(value) => onLimitChange(Number(value))}>
            <SelectTrigger className="w-20" aria-labelledby={pageSizeId}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {pageSizeOptions.map((option) => (
                <SelectItem key={option} value={String(option)}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            aria-label="Previous page"
            disabled={!canGoBack}
            onClick={() => onOffsetChange(Math.max(0, offset - limit))}
          >
            <ChevronLeft />
          </Button>
          <Button
            variant="outline"
            size="icon"
            aria-label="Next page"
            disabled={!canGoForward}
            onClick={() => onOffsetChange(offset + limit)}
          >
            <ChevronRight />
          </Button>
        </div>
      </div>
    </div>
  );
}

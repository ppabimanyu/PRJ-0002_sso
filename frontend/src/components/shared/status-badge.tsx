import type { ReactNode } from "react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type Status = "success" | "warning" | "danger" | "neutral" | "pending";

type StatusStyle = {
  badge: string;
  dot: string;
};

const statusStyles: Record<Status, StatusStyle> = {
  success: {
    badge: "border-success/30 bg-success/10 text-success",
    dot: "bg-success",
  },
  warning: {
    badge: "border-warning/30 bg-warning/10 text-warning",
    dot: "bg-warning",
  },
  danger: {
    badge: "border-destructive/30 bg-destructive/10 text-destructive",
    dot: "bg-destructive",
  },
  neutral: {
    badge: "border-border bg-muted text-muted-foreground",
    dot: "bg-muted-foreground",
  },
  pending: {
    badge: "border-border bg-muted text-muted-foreground",
    dot: "animate-pulse bg-muted-foreground",
  },
};

type StatusBadgeProps = {
  status: Status;
  label: ReactNode;
  dot?: boolean;
  className?: string;
};

export function StatusBadge({ status, label, dot = true, className }: StatusBadgeProps) {
  const style = statusStyles[status];

  return (
    <Badge variant="outline" className={cn("gap-1.5", style.badge, className)}>
      {dot ? <span className={cn("size-1.5 rounded-full", style.dot)} aria-hidden="true" /> : null}
      {label}
    </Badge>
  );
}

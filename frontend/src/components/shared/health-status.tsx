import { RefreshCw } from "lucide-react";

import { StatusBadge, type Status } from "@/components/shared/status-badge";
import { Button } from "@/components/ui/button";
import { useHealth } from "@/hooks/use-health";

type HealthState = "checking" | "operational" | "degraded";

const healthStates: Record<HealthState, { status: Status; label: string }> = {
  checking: { status: "pending", label: "Checking" },
  operational: { status: "success", label: "Operational" },
  degraded: { status: "danger", label: "Degraded" },
};

export function HealthStatus() {
  const { data, isError, isFetching, isPending, refetch } = useHealth();

  const state: HealthState = isPending
    ? "checking"
    : isError || data?.status !== "ok"
      ? "degraded"
      : "operational";

  const { status, label } = healthStates[state];
  const checks = Object.entries(data?.checks ?? {});

  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex flex-wrap items-center gap-x-6 gap-y-3">
        <span className="text-sm font-medium">Readiness</span>
        <StatusBadge status={status} label={label} />
        {checks.length > 0 ? (
          <dl className="flex flex-wrap items-center gap-x-5 gap-y-1 text-sm">
            {checks.map(([name, value]) => (
              <div key={name} className="flex items-center gap-1.5">
                <dt className="first-letter:uppercase text-muted-foreground">{name}</dt>
                <dd className={value === "ok" ? "text-foreground" : "text-destructive"}>
                  {value === "ok" ? "responding" : "not responding"}
                </dd>
              </div>
            ))}
          </dl>
        ) : (
          <span className="text-sm text-muted-foreground">
            {isError ? "The backend did not answer. Start the stack with make dev." : null}
          </span>
        )}
      </div>
      <Button
        variant="ghost"
        size="sm"
        className="-ml-2 self-start sm:ml-0 sm:self-auto"
        onClick={() => void refetch()}
        disabled={isFetching}
      >
        <RefreshCw className={isFetching ? "animate-spin" : undefined} />
        Check again
      </Button>
    </div>
  );
}

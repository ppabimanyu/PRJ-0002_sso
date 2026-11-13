import { cn } from "@/lib/utils";

import { appName, chartCells, serviceCell } from "./brand";

/**
 * The house test chart: a fixed grid of cells where the service occupies one.
 * Same chart in every project, different lit cell, so the identity stays
 * recognisable while a service remains distinguishable at a glance.
 */
export function ServiceChart({ name = appName, className }: { name?: string; className?: string }) {
  const active = serviceCell(name);

  return (
    <div
      aria-hidden="true"
      data-slot="service-chart"
      className={cn("grid grid-cols-6 gap-2", className)}
    >
      {Array.from({ length: chartCells }, (_, index) => (
        <span
          key={index}
          data-active={index === active ? "true" : undefined}
          className={cn(
            "aspect-square border",
            index === active ? "border-brand bg-brand" : "border-border bg-muted",
          )}
        />
      ))}
    </div>
  );
}

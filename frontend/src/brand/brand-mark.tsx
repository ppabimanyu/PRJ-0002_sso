import { cn } from "@/lib/utils";

import { appName } from "./brand";

/**
 * The house mark: a registration mark. Four corner brackets align a sheet;
 * the solid core is the service being placed. It reads at 16px and at 400px.
 */
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
      data-slot="brand-mark"
      className={cn("size-5", className)}
    >
      <g stroke="currentColor" strokeWidth="1.6">
        <path d="M2 8V2h6" />
        <path d="M22 8V2h-6" />
        <path d="M2 16v6h6" />
        <path d="M22 16v6h-6" />
      </g>
      <rect x="9.5" y="9.5" width="5" height="5" fill="currentColor" />
    </svg>
  );
}

export function BrandLockup({ className }: { className?: string }) {
  return (
    <span className={cn("flex items-center gap-2.5", className)}>
      <BrandMark className="size-5 text-brand" />
      <span className="text-sm font-semibold tracking-tight">{appName}</span>
    </span>
  );
}

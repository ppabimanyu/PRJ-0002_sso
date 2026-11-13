import { cn } from "@/lib/utils";

/**
 * Crop marks at the corners of a region. They frame a page as one sheet from
 * the same press, and stay out of the reading flow.
 */
export function SheetCorners({
  className,
  cornerClassName,
}: {
  className?: string;
  cornerClassName?: string;
}) {
  const corner = cn("absolute border-brand/60", cornerClassName ?? "size-3");

  return (
    <div aria-hidden="true" className={cn("pointer-events-none absolute inset-0", className)}>
      <span className={cn(corner, "top-0 left-0 border-t border-l")} />
      <span className={cn(corner, "top-0 right-0 border-t border-r")} />
      <span className={cn(corner, "bottom-0 left-0 border-b border-l")} />
      <span className={cn(corner, "right-0 bottom-0 border-r border-b")} />
    </div>
  );
}

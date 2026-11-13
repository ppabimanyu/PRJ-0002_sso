import { Spinner } from "@/components/ui/spinner";

export function RouteFallback() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Spinner className="size-5" aria-label="Loading" />
    </div>
  );
}

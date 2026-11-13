import { Link } from "react-router-dom";

import { BrandMark } from "@/brand/brand-mark";
import { Button } from "@/components/ui/button";

export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center gap-6 py-20 text-center">
      <BrandMark className="size-8 text-brand" />
      <div className="flex flex-col items-center gap-2">
        <p className="text-sm text-muted-foreground tabular-nums">404</p>
        <h1 className="text-2xl font-semibold tracking-tight">Page not found</h1>
        <p className="max-w-[46ch] text-sm leading-relaxed text-muted-foreground">
          This address does not exist in the service. Check the link, or start again from the home
          page.
        </p>
      </div>
      <Button asChild>
        <Link to="/">Go home</Link>
      </Button>
    </div>
  );
}

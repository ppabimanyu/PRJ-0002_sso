import { isRouteErrorResponse, Link, useRouteError } from "react-router-dom";

import { StatusPage } from "@/components/shared/status-page";
import { Button } from "@/components/ui/button";

function describe(error: unknown): string {
  if (isRouteErrorResponse(error)) {
    return `${error.status} ${error.statusText}`.trim();
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return "The page could not be rendered.";
}

export function RouteErrorPage() {
  const error = useRouteError();
  const notFound = isRouteErrorResponse(error) && error.status === 404;

  return (
    <StatusPage
      title={notFound ? "Page not found" : "Something went wrong"}
      description={describe(error)}
      detail={
        import.meta.env.DEV && !notFound ? (
          <pre className="max-h-40 overflow-auto rounded-lg bg-muted p-3 text-xs whitespace-pre-wrap">
            {error instanceof Error ? (error.stack ?? error.message) : describe(error)}
          </pre>
        ) : null
      }
      actions={
        <>
          <Button onClick={() => window.location.reload()}>Reload page</Button>
          <Button variant="ghost" asChild>
            <Link to="/">Go home</Link>
          </Button>
        </>
      }
    />
  );
}

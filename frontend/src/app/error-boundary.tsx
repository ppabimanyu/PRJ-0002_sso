import { Component, type ErrorInfo, type ReactNode } from "react";

import { StatusPage } from "@/components/shared/status-page";
import { Button } from "@/components/ui/button";

type ErrorBoundaryProps = {
  children: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
};

/**
 * Catches failures outside route rendering, such as providers or the router
 * failing to initialise. Route render errors are handled by the router through
 * the root route's error element.
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Unhandled render error", error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      return (
        <StatusPage
          title="Something went wrong"
          description="The page could not be rendered. Reloading usually fixes it."
          actions={<Button onClick={() => window.location.reload()}>Reload page</Button>}
        />
      );
    }

    return this.props.children;
  }
}

import type { ReactNode } from "react";

import { MaintenancePage } from "./maintenance-page";

/**
 * Replaces the router with the maintenance page when the build was made with
 * `VITE_MAINTENANCE_MODE=true`. A build-time flag is deliberate: the backend is
 * usually down during a maintenance window, so it cannot be the source of truth.
 */
export function MaintenanceGate({ children }: { children: ReactNode }) {
  if (import.meta.env.VITE_MAINTENANCE_MODE === "true") {
    return <MaintenancePage />;
  }

  return <>{children}</>;
}

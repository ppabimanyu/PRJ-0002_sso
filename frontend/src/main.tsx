import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router-dom";

import { ErrorBoundary } from "./app/error-boundary";
import { MaintenanceGate } from "./app/maintenance-gate";
import { Providers } from "./app/providers";
import { router } from "./app/router";
import { appName } from "./brand/brand";
import "./index.css";

document.title = appName;

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Providers>
      <ErrorBoundary>
        <MaintenanceGate>
          <RouterProvider router={router} />
        </MaintenanceGate>
      </ErrorBoundary>
    </Providers>
  </StrictMode>,
);

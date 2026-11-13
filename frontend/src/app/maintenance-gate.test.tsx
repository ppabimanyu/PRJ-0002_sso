import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { MaintenanceGate } from "./maintenance-gate";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("MaintenanceGate", () => {
  it("renders the maintenance page when the build flag is on", () => {
    vi.stubEnv("VITE_MAINTENANCE_MODE", "true");

    render(
      <MaintenanceGate>
        <p>Application</p>
      </MaintenanceGate>,
    );

    expect(screen.getByText("Down for maintenance")).toBeInTheDocument();
    expect(screen.queryByText("Application")).not.toBeInTheDocument();
  });

  it("renders the application when the flag is off", () => {
    vi.stubEnv("VITE_MAINTENANCE_MODE", "false");

    render(
      <MaintenanceGate>
        <p>Application</p>
      </MaintenanceGate>,
    );

    expect(screen.getByText("Application")).toBeInTheDocument();
    expect(screen.queryByText("Down for maintenance")).not.toBeInTheDocument();
  });
});

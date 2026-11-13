import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { server } from "@/test/setup";
import { HealthStatus } from "./health-status";

function renderHealthStatus() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <HealthStatus />
    </QueryClientProvider>,
  );
}

describe("HealthStatus", () => {
  it("shows operational when readiness is ok", async () => {
    server.use(
      http.get("*/readyz", () =>
        HttpResponse.json({ status: "ok", checks: { postgres: "ok", redis: "ok" } }),
      ),
    );

    renderHealthStatus();

    expect(await screen.findByText("Operational")).toBeInTheDocument();
  });

  it("shows degraded when readiness reports a degraded status", async () => {
    server.use(
      http.get("*/readyz", () =>
        HttpResponse.json({ status: "degraded", checks: { postgres: "error", redis: "ok" } }),
      ),
    );

    renderHealthStatus();

    expect(await screen.findByText("Degraded")).toBeInTheDocument();
  });
});

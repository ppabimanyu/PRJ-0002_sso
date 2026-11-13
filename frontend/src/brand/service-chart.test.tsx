import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { chartCells, serviceCell } from "./brand";
import { ServiceChart } from "./service-chart";

describe("serviceCell", () => {
  it("is deterministic for the same name", () => {
    expect(serviceCell("Billing")).toBe(serviceCell("Billing"));
  });

  it("stays inside the chart", () => {
    const cell = serviceCell("Billing");

    expect(cell).toBeGreaterThanOrEqual(0);
    expect(cell).toBeLessThan(chartCells);
  });
});

describe("ServiceChart", () => {
  it("lights exactly one cell", () => {
    const { container } = render(<ServiceChart name="Billing" />);

    const cells = container.querySelectorAll("[data-slot='service-chart'] > span");
    const active = container.querySelectorAll("[data-active='true']");

    expect(cells).toHaveLength(chartCells);
    expect(active).toHaveLength(1);
  });

  it("puts the lit cell where the name says", () => {
    const { container } = render(<ServiceChart name="Billing" />);

    const cells = [...container.querySelectorAll("[data-slot='service-chart'] > span")];
    const index = cells.findIndex((cell) => cell.getAttribute("data-active") === "true");

    expect(index).toBe(serviceCell("Billing"));
  });
});

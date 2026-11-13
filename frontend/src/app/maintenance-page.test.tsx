import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { MaintenancePage } from "./maintenance-page";

describe("MaintenancePage", () => {
  it("explains the window and offers a retry", () => {
    render(<MaintenancePage />);

    expect(screen.getByText("Down for maintenance")).toBeInTheDocument();
    expect(screen.getByText(/will be back shortly/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument();
  });
});

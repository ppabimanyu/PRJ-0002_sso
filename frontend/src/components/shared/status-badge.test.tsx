import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StatusBadge, type Status } from "./status-badge";

const statuses: Status[] = ["success", "warning", "danger", "neutral", "pending"];

describe("StatusBadge", () => {
  it.each(statuses)("renders the %s label", (status) => {
    render(<StatusBadge status={status} label={status} />);

    expect(screen.getByText(status)).toBeInTheDocument();
  });

  it("hides the dot from assistive technology", () => {
    const { container } = render(<StatusBadge status="success" label="Operational" />);

    expect(container.querySelector("[aria-hidden='true']")).toBeInTheDocument();
  });

  it("omits the dot when asked", () => {
    const { container } = render(<StatusBadge status="success" label="Operational" dot={false} />);

    expect(container.querySelector("[aria-hidden='true']")).not.toBeInTheDocument();
  });
});

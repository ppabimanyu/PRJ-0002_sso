import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Pagination } from "./pagination";

Element.prototype.hasPointerCapture = vi.fn();
Element.prototype.releasePointerCapture = vi.fn();
Element.prototype.scrollIntoView = vi.fn();

function renderPagination(overrides: Partial<Parameters<typeof Pagination>[0]> = {}) {
  const props = {
    total: 134,
    limit: 20,
    offset: 0,
    onOffsetChange: vi.fn(),
    onLimitChange: vi.fn(),
    ...overrides,
  };

  render(<Pagination {...props} />);

  return props;
}

describe("Pagination", () => {
  it("shows the visible range", () => {
    renderPagination();

    expect(screen.getByText("1–20 of 134")).toBeInTheDocument();
  });

  it("shows the last page range and disables next", () => {
    renderPagination({ offset: 120 });

    expect(screen.getByText("121–134 of 134")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next page" })).toBeDisabled();
  });

  it("disables previous on the first page", () => {
    renderPagination();

    expect(screen.getByRole("button", { name: "Previous page" })).toBeDisabled();
  });

  it("moves one page forward", async () => {
    const props = renderPagination();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Next page" }));

    expect(props.onOffsetChange).toHaveBeenCalledWith(20);
  });

  it("reports the new page size and resets the offset to zero", async () => {
    const props = renderPagination({ offset: 60 });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox", { name: "Rows per page" }));
    await user.click(screen.getByRole("option", { name: "50" }));

    expect(props.onLimitChange).toHaveBeenCalledWith(50);
  });

  it("shows a zero range when there are no rows", () => {
    renderPagination({ total: 0 });

    expect(screen.getByText("0–0 of 0")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next page" })).toBeDisabled();
  });
});

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ConfirmDialog } from "./confirm-dialog";

function renderDialog(overrides: Partial<Parameters<typeof ConfirmDialog>[0]> = {}) {
  const props = {
    open: true,
    onOpenChange: vi.fn(),
    onConfirm: vi.fn(),
    title: "Delete this project?",
    description: "This action cannot be undone.",
    confirmLabel: "Delete",
    destructive: true,
    ...overrides,
  };

  render(<ConfirmDialog {...props} />);

  return props;
}

describe("ConfirmDialog", () => {
  it("shows the question and the consequence", () => {
    renderDialog();

    expect(screen.getByText("Delete this project?")).toBeInTheDocument();
    expect(screen.getByText("This action cannot be undone.")).toBeInTheDocument();
  });

  it("calls onConfirm", async () => {
    const props = renderDialog();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Delete" }));

    expect(props.onConfirm).toHaveBeenCalled();
  });

  it("reports the close request on cancel", async () => {
    const props = renderDialog();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(props.onOpenChange).toHaveBeenCalledWith(false);
  });

  it("renders nothing while closed", () => {
    renderDialog({ open: false });

    expect(screen.queryByText("Delete this project?")).not.toBeInTheDocument();
  });
});

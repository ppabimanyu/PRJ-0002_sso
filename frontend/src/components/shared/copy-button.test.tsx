import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CopyButton } from "./copy-button";

function stubClipboard(writeText: ReturnType<typeof vi.fn>) {
  Object.defineProperty(navigator, "clipboard", {
    value: { writeText },
    configurable: true,
  });
}

describe("CopyButton", () => {
  it("copies the value and confirms", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    stubClipboard(writeText);

    render(<CopyButton value="npm run dev" label="Copy command" />);
    await user.click(screen.getByRole("button", { name: "Copy command" }));

    await waitFor(() => expect(writeText).toHaveBeenCalledWith("npm run dev"));
    expect(await screen.findByRole("button", { name: "Copied" })).toBeInTheDocument();
  });

  it("stays idle when the clipboard rejects", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockRejectedValue(new Error("denied"));
    stubClipboard(writeText);

    render(<CopyButton value="npm run dev" />);
    await user.click(screen.getByRole("button", { name: "Copy" }));

    await waitFor(() => expect(writeText).toHaveBeenCalled());
    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
  });

  it("degrades when the clipboard is unavailable", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn();
    Object.defineProperty(navigator, "clipboard", { value: undefined, configurable: true });

    render(<CopyButton value="npm run dev" />);
    await user.click(screen.getByRole("button", { name: "Copy" }));

    expect(writeText).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
  });
});

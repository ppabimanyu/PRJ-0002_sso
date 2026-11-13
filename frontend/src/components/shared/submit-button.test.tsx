import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SubmitButton } from "./submit-button";

describe("SubmitButton", () => {
  it("renders its label and submits", () => {
    render(<SubmitButton>Save changes</SubmitButton>);

    const button = screen.getByRole("button", { name: "Save changes" });
    expect(button).toHaveAttribute("type", "submit");
    expect(button).toBeEnabled();
  });

  it("switches to the pending label and disables itself while submitting", () => {
    render(
      <SubmitButton isSubmitting pendingLabel="Saving">
        Save changes
      </SubmitButton>,
    );

    const button = screen.getByRole("button", { name: /Saving/ });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("aria-busy", "true");
    expect(screen.queryByText("Save changes")).not.toBeInTheDocument();
  });
});

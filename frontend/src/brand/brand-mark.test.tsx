import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { appName } from "./brand";
import { BrandLockup, BrandMark } from "./brand-mark";

describe("BrandMark", () => {
  it("renders a decorative mark instead of an image with alt text", () => {
    const { container } = render(<BrandMark />);

    const mark = container.querySelector("[data-slot='brand-mark']");
    expect(mark).toBeInTheDocument();
    expect(mark).toHaveAttribute("aria-hidden", "true");
  });
});

describe("BrandLockup", () => {
  it("shows the configured app name next to the mark", () => {
    render(<BrandLockup />);

    expect(screen.getByText(appName)).toBeInTheDocument();
  });
});

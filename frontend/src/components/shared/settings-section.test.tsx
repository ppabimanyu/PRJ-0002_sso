import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SettingsSection } from "./settings-section";

describe("SettingsSection", () => {
  it("renders its heading, description, action, and content", () => {
    render(
      <SettingsSection
        title="Profile"
        description="How your account appears."
        actions={<button type="button">Save</button>}
      >
        <p>Display name</p>
      </SettingsSection>,
    );

    expect(screen.getByText("Profile")).toBeInTheDocument();
    expect(screen.getByText("How your account appears.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
    expect(screen.getByText("Display name")).toBeInTheDocument();
  });

  it("omits the description and action when they are not given", () => {
    render(
      <SettingsSection title="Notifications">
        <p>Email digest</p>
      </SettingsSection>,
    );

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});

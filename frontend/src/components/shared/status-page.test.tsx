import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StatusPage } from "./status-page";

describe("StatusPage", () => {
  it("renders the title, description, detail, and actions", () => {
    render(
      <StatusPage
        icon={<span data-testid="icon" />}
        title="Down for maintenance"
        description="We will be back shortly."
        detail={<p>Expected back at 12:00 UTC</p>}
        actions={<button type="button">Try again</button>}
      />,
    );

    expect(screen.getByTestId("icon")).toBeInTheDocument();
    expect(screen.getByText("Down for maintenance")).toBeInTheDocument();
    expect(screen.getByText("We will be back shortly.")).toBeInTheDocument();
    expect(screen.getByText("Expected back at 12:00 UTC")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument();
  });

  it("omits the optional parts", () => {
    render(<StatusPage title="Something went wrong" />);

    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});

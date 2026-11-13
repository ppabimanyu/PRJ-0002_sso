import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DataState } from "./data-state";

describe("DataState", () => {
  it("announces the loading state", () => {
    render(<DataState status="loading" />);

    expect(screen.getByRole("status")).toHaveAttribute("aria-label", "Loading");
    expect(screen.getByText("Loading")).toBeInTheDocument();
  });

  it("shows the error message and retries", async () => {
    const onRetry = vi.fn();

    render(<DataState status="error" error={new Error("Backend unavailable")} onRetry={onRetry} />);

    expect(screen.getByRole("alert")).toHaveTextContent("Backend unavailable");

    await userEvent.click(screen.getByRole("button", { name: "Try again" }));

    expect(onRetry).toHaveBeenCalledOnce();
  });

  it("renders the empty state with an action", () => {
    render(
      <DataState
        status="empty"
        title="No projects"
        description="Create your first project."
        action={<button type="button">New project</button>}
      />,
    );

    expect(screen.getByText("No projects")).toBeInTheDocument();
    expect(screen.getByText("Create your first project.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "New project" })).toBeInTheDocument();
  });

  it("renders ready children", () => {
    render(
      <DataState status="ready">
        <p>Project list</p>
      </DataState>,
    );

    expect(screen.getByText("Project list")).toBeInTheDocument();
  });
});

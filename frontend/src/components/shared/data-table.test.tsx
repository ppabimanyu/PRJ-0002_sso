import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DataTable, type DataTableColumn } from "./data-table";

type Project = {
  id: string;
  name: string;
  stars: number;
};

const rows: Project[] = [
  { id: "1", name: "Charlie", stars: 3 },
  { id: "2", name: "Alpha", stars: 10 },
  { id: "3", name: "Bravo", stars: 7 },
];

const columns: DataTableColumn<Project>[] = [
  { id: "name", header: "Project", accessor: (row) => row.name, sortable: true },
  { id: "stars", header: "Stars", accessor: (row) => row.stars, sortable: true },
];

function renderTable(overrides: Partial<Parameters<typeof DataTable<Project>>[0]> = {}) {
  const props = {
    columns,
    rows,
    rowKey: (row: Project) => row.id,
    total: 134,
    limit: 20,
    offset: 0,
    onOffsetChange: vi.fn(),
    onLimitChange: vi.fn(),
    ...overrides,
  };

  render(<DataTable<Project> {...props} />);

  return props;
}

function bodyRows() {
  return screen.getAllByRole("row").slice(1);
}

describe("DataTable", () => {
  it("renders each row and the range", () => {
    renderTable();

    expect(bodyRows()).toHaveLength(3);
    expect(screen.getByText("Alpha")).toBeInTheDocument();
    expect(screen.getByText("1–20 of 134")).toBeInTheDocument();
  });

  it("sorts the current page ascending then descending then back", async () => {
    renderTable();
    const user = userEvent.setup();
    const header = screen.getByRole("button", { name: "Project" });

    await user.click(header);
    expect(bodyRows()[0]).toHaveTextContent("Alpha");
    expect(header.closest("th")).toHaveAttribute("aria-sort", "ascending");

    await user.click(header);
    expect(bodyRows()[0]).toHaveTextContent("Charlie");
    expect(header.closest("th")).toHaveAttribute("aria-sort", "descending");

    await user.click(header);
    expect(bodyRows()[0]).toHaveTextContent("Charlie");
    expect(header.closest("th")).toHaveAttribute("aria-sort", "none");
  });

  it("sorts numbers numerically", async () => {
    renderTable();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Stars" }));

    expect(within(bodyRows()[0]).getByText("3")).toBeInTheDocument();
    expect(bodyRows()[2]).toHaveTextContent("10");
  });

  it("renders a custom cell", () => {
    renderTable({
      columns: [
        {
          id: "name",
          header: "Project",
          accessor: (row) => row.name,
          cell: (row) => <strong>{row.name.toUpperCase()}</strong>,
        },
      ],
    });

    expect(screen.getByText("ALPHA")).toBeInTheDocument();
  });

  it("delegates loading to the data state", () => {
    renderTable({ isLoading: true });

    expect(screen.getByText("Loading")).toBeInTheDocument();
  });

  it("delegates errors with a retry action", async () => {
    const onRetry = vi.fn();
    renderTable({ error: new Error("The projects service did not respond."), onRetry });
    const user = userEvent.setup();

    expect(screen.getByText("The projects service did not respond.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Try again" }));

    expect(onRetry).toHaveBeenCalled();
  });

  it("keeps pagination reachable when the current page is empty", () => {
    renderTable({ rows: [], total: 134, offset: 200, emptyTitle: "No projects on this page" });

    expect(screen.getByText("No projects on this page")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous page" })).toBeEnabled();
  });

  it("hides pagination when the result set is empty", () => {
    renderTable({ rows: [], total: 0 });

    expect(screen.getByText("No data")).toBeInTheDocument();
    expect(screen.queryByText(/of 0/)).not.toBeInTheDocument();
  });
});

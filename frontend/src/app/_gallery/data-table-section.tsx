import { useState } from "react";

import { DataTable, type DataTableColumn } from "@/components/shared/data-table";
import { StatusBadge, type Status } from "@/components/shared/status-badge";
import { useListQuery } from "@/hooks/use-list-query";
import { GallerySection } from "./section";

type Project = {
  id: string;
  name: string;
  owner: string;
  status: Status;
  updated: string;
};

const owners = ["Ada Lovelace", "Grace Hopper", "Alan Turing"];
const statuses: Status[] = ["success", "pending", "warning"];
const statusLabels: Record<Status, string> = {
  success: "Healthy",
  warning: "At risk",
  danger: "Failed",
  neutral: "Draft",
  pending: "Building",
};

const projects: Project[] = Array.from({ length: 134 }, (_, index) => ({
  id: `project-${index + 1}`,
  name: `Project ${String(index + 1).padStart(3, "0")}`,
  owner: owners[index % owners.length],
  status: statuses[index % statuses.length],
  updated: `${index + 1} days ago`,
}));

const columns: DataTableColumn<Project>[] = [
  { id: "name", header: "Project", accessor: (row) => row.name, sortable: true },
  { id: "owner", header: "Owner", accessor: (row) => row.owner, sortable: true },
  {
    id: "status",
    header: "Status",
    accessor: (row) => row.status,
    cell: (row) => <StatusBadge status={row.status} label={statusLabels[row.status]} />,
  },
  {
    id: "updated",
    header: "Updated",
    accessor: (row) => row.updated,
    className: "text-muted-foreground",
  },
];

export function DataTableSection() {
  const { limit, offset, setLimit, setOffset } = useListQuery();
  const [error, setError] = useState<Error | null>(null);

  const rows = projects.slice(offset, offset + limit);

  return (
    <GallerySection
      title="Data table"
      description="Offset pagination over a fake 134-row set. Sorting reorders the current page only, because the API has no sort parameter yet."
    >
      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          className="text-sm text-muted-foreground underline underline-offset-4"
          onClick={() =>
            setError(error ? null : new Error("The projects service did not respond."))
          }
        >
          {error ? "Show rows again" : "Show the error state"}
        </button>
      </div>
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(row) => row.id}
        total={projects.length}
        limit={limit}
        offset={offset}
        onOffsetChange={setOffset}
        onLimitChange={setLimit}
        error={error}
        onRetry={() => setError(null)}
        emptyTitle="No projects on this page"
        emptyDescription="Go back a page, or lower the offset in the URL."
      />
    </GallerySection>
  );
}

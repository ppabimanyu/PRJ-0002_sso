import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";

import { DataState } from "@/components/shared/data-state";
import { Pagination } from "@/components/shared/pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export type DataTableColumn<T> = {
  id: string;
  header: ReactNode;
  accessor: (row: T) => unknown;
  cell?: (row: T) => ReactNode;
  sortable?: boolean;
  className?: string;
};

type SortState = {
  id: string;
  direction: "asc" | "desc";
};

type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  total: number;
  limit: number;
  offset: number;
  onOffsetChange: (offset: number) => void;
  onLimitChange: (limit: number) => void;
  isLoading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  emptyTitle?: ReactNode;
  emptyDescription?: ReactNode;
};

function compareValues(a: unknown, b: unknown): number {
  if (typeof a === "number" && typeof b === "number") {
    return a - b;
  }

  return String(a ?? "").localeCompare(String(b ?? ""), undefined, {
    numeric: true,
    sensitivity: "base",
  });
}

/**
 * Sorts only the rows of the current page. The API has no `sort` parameter yet,
 * so a sorted column is not a sorted result set. Revisit when the backend
 * contract grows sorting.
 */
export function DataTable<T>({
  columns,
  rows,
  rowKey,
  total,
  limit,
  offset,
  onOffsetChange,
  onLimitChange,
  isLoading = false,
  error = null,
  onRetry,
  emptyTitle,
  emptyDescription,
}: DataTableProps<T>) {
  const [sort, setSort] = useState<SortState | null>(null);

  const sortedRows = useMemo(() => {
    if (!sort) {
      return rows;
    }

    const column = columns.find((candidate) => candidate.id === sort.id);

    if (!column) {
      return rows;
    }

    const factor = sort.direction === "asc" ? 1 : -1;

    return [...rows].sort(
      (left, right) => compareValues(column.accessor(left), column.accessor(right)) * factor,
    );
  }, [rows, columns, sort]);

  const pagination = (
    <Pagination
      total={total}
      limit={limit}
      offset={offset}
      onOffsetChange={onOffsetChange}
      onLimitChange={onLimitChange}
    />
  );

  if (isLoading) {
    return <DataState status="loading" />;
  }

  if (error) {
    return <DataState status="error" error={error} onRetry={onRetry} />;
  }

  if (rows.length === 0) {
    return (
      <div className="flex flex-col gap-3">
        <DataState status="empty" title={emptyTitle} description={emptyDescription} />
        {total > 0 ? pagination : null}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="overflow-hidden rounded-2xl border">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((column) => {
                const activeSort = sort?.id === column.id ? sort : null;
                const ariaSort: "none" | "ascending" | "descending" = activeSort
                  ? activeSort.direction === "asc"
                    ? "ascending"
                    : "descending"
                  : "none";

                return (
                  <TableHead
                    key={column.id}
                    aria-sort={column.sortable ? ariaSort : undefined}
                    className={column.className}
                  >
                    {column.sortable ? (
                      <button
                        type="button"
                        className="flex items-center gap-1.5 font-medium"
                        onClick={() =>
                          setSort((previous) => {
                            if (previous?.id !== column.id) {
                              return { id: column.id, direction: "asc" };
                            }

                            return previous.direction === "asc"
                              ? { id: column.id, direction: "desc" }
                              : null;
                          })
                        }
                      >
                        {column.header}
                        {activeSort ? (
                          activeSort.direction === "asc" ? (
                            <ArrowUp className="size-3.5" aria-hidden="true" />
                          ) : (
                            <ArrowDown className="size-3.5" aria-hidden="true" />
                          )
                        ) : (
                          <ArrowUpDown
                            className="size-3.5 text-muted-foreground"
                            aria-hidden="true"
                          />
                        )}
                      </button>
                    ) : (
                      column.header
                    )}
                  </TableHead>
                );
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedRows.map((row) => (
              <TableRow key={rowKey(row)}>
                {columns.map((column) => (
                  <TableCell key={column.id} className={column.className}>
                    {column.cell ? column.cell(row) : String(column.accessor(row) ?? "")}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      {pagination}
    </div>
  );
}

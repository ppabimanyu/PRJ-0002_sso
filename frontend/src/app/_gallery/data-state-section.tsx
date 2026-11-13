import { useState } from "react";
import { toast } from "sonner";

import { DataState } from "@/components/shared/data-state";
import { Button } from "@/components/ui/button";
import { GalleryDemo, GallerySection } from "./section";

const dataStateError = new Error("The projects service did not respond.");

export function DataStateSection() {
  const [dataStatus, setDataStatus] = useState<"error" | "ready">("error");

  return (
    <GallerySection title="Data state" description="One component, four mutually exclusive states.">
      <div className="grid gap-8 md:grid-cols-2">
        <GalleryDemo label="Loading">
          <DataState status="loading" />
        </GalleryDemo>
        <GalleryDemo label="Empty">
          <DataState
            status="empty"
            title="No projects yet"
            description="Create a project to see it here."
            action={
              <Button type="button" size="sm" onClick={() => toast("This is a gallery example.")}>
                New project
              </Button>
            }
          />
        </GalleryDemo>
        <GalleryDemo label="Error">
          {dataStatus === "error" ? (
            <DataState
              status="error"
              error={dataStateError}
              onRetry={() => setDataStatus("ready")}
            />
          ) : (
            <DataState status="ready">
              <div className="rounded-2xl border bg-card p-4 text-sm">
                <p className="font-medium">The service is back</p>
                <p className="text-muted-foreground">The retry action switched this example.</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="mt-3"
                  onClick={() => setDataStatus("error")}
                >
                  Show error again
                </Button>
              </div>
            </DataState>
          )}
        </GalleryDemo>
        <GalleryDemo label="Ready">
          <DataState status="ready">
            <div className="rounded-2xl border bg-card p-4 text-sm">
              <p className="font-medium">3 projects</p>
              <p className="text-muted-foreground">Ready content renders as children.</p>
            </div>
          </DataState>
        </GalleryDemo>
      </div>
    </GallerySection>
  );
}

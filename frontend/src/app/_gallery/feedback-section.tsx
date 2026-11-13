import { useState } from "react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { CopyButton } from "@/components/shared/copy-button";
import { Button } from "@/components/ui/button";
import { GalleryDemo, GallerySection } from "./section";

export function FeedbackSection() {
  const [open, setOpen] = useState(false);

  return (
    <GallerySection
      title="Feedback"
      description="Destructive confirmation stays explicit at the call site, and copying confirms with a toast."
    >
      <GalleryDemo label="Confirm dialog">
        <div>
          <Button variant="destructive" onClick={() => setOpen(true)}>
            Delete project
          </Button>
        </div>
        <ConfirmDialog
          open={open}
          onOpenChange={setOpen}
          onConfirm={() => {
            setOpen(false);
            toast.success("Project deleted");
          }}
          title="Delete this project?"
          description="The project and its data are removed. This cannot be undone."
          confirmLabel="Delete"
          destructive
        />
      </GalleryDemo>

      <GalleryDemo label="Copy button">
        <div>
          <CopyButton value="npm run dev" label="Copy command" />
        </div>
      </GalleryDemo>
    </GallerySection>
  );
}

import { useState, type FormEvent } from "react";

import { SearchInput } from "@/components/shared/search-input";
import { SubmitButton } from "@/components/shared/submit-button";
import { GalleryDemo, GallerySection } from "./section";

export function FormsSection() {
  const [query, setQuery] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    window.setTimeout(() => setIsSubmitting(false), 1500);
  }

  return (
    <GallerySection
      title="Form controls"
      description="Search waits for a pause; the submit button reads pending state from the form that owns it."
    >
      <GalleryDemo label="Search input">
        <SearchInput value={query} onValueChange={setQuery} placeholder="Search projects" />
        <p className="text-sm text-muted-foreground">Reported query: {query || "—"}</p>
      </GalleryDemo>

      <GalleryDemo label="Submit button">
        <form onSubmit={handleSubmit} className="flex flex-wrap items-center gap-3">
          <SubmitButton isSubmitting={isSubmitting} pendingLabel="Saving">
            Save changes
          </SubmitButton>
          <p className="text-sm text-muted-foreground">
            The pending state comes from the form, not from a context.
          </p>
        </form>
      </GalleryDemo>
    </GallerySection>
  );
}

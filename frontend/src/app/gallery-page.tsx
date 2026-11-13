import { PageHeader } from "@/components/shared/page-header";
import { Separator } from "@/components/ui/separator";
import { DataStateSection } from "./_gallery/data-state-section";
import { DataTableSection } from "./_gallery/data-table-section";
import { DisplaySection } from "./_gallery/display-section";
import { FeedbackSection } from "./_gallery/feedback-section";
import { FormFieldSection } from "./_gallery/form-field-section";
import { FormsSection } from "./_gallery/forms-section";
import { HouseStyleSection } from "./_gallery/house-style-section";
import { OverlaySection } from "./_gallery/overlay-section";
import { SelectionSection } from "./_gallery/selection-section";
import { SettingsSectionDemo } from "./_gallery/settings-demo-section";

export function GalleryPage() {
  return (
    <div className="flex flex-col gap-10">
      <PageHeader
        title="Component gallery"
        description="Dev-only examples for the primitives and shared components. This route is not part of the production build."
      />

      <HouseStyleSection />
      <Separator />
      <DataStateSection />
      <Separator />
      <FormFieldSection />
      <Separator />
      <OverlaySection />
      <Separator />
      <SelectionSection />
      <Separator />
      <DisplaySection />
      <Separator />
      <DataTableSection />
      <Separator />
      <FormsSection />
      <Separator />
      <FeedbackSection />
      <Separator />
      <SettingsSectionDemo />
    </div>
  );
}

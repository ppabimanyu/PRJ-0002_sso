import { PageHeader } from "@/components/shared/page-header";
import { SettingsSection } from "@/components/shared/settings-section";

export function AdminPage() {
  return (
    <div className="flex flex-col gap-8">
      <PageHeader
        title="Admin"
        description="Manage the Instance, its Projects, and the people who can access them."
      />

      <SettingsSection
        title="Instance administration"
        description="The admin console arrives in the next milestone."
      >
        <p className="text-sm leading-relaxed text-muted-foreground">
          Project, Application, Role, Membership, and User management will appear here.
        </p>
      </SettingsSection>
    </div>
  );
}

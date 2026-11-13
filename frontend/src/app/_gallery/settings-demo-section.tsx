import { SettingsSection } from "@/components/shared/settings-section";
import { Button } from "@/components/ui/button";
import { GallerySection } from "./section";

export function SettingsSectionDemo() {
  return (
    <GallerySection
      title="Settings section"
      description="A card header with an optional action, used to stack settings on one page."
    >
      <SettingsSection
        title="Profile"
        description="How your account appears to other people."
        actions={
          <Button size="sm" type="button">
            Save
          </Button>
        }
      >
        <div className="flex flex-col gap-1">
          <p className="text-sm font-medium">Display name</p>
          <p className="text-sm text-muted-foreground">Ada Lovelace</p>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-sm font-medium">Email</p>
          <p className="text-sm text-muted-foreground">ada@example.com</p>
        </div>
      </SettingsSection>

      <SettingsSection title="Notifications" description="What we email you about.">
        <div className="flex flex-col gap-1">
          <p className="text-sm font-medium">Weekly digest</p>
          <p className="text-sm text-muted-foreground">Every Monday at 08:00 UTC.</p>
        </div>
      </SettingsSection>
    </GallerySection>
  );
}

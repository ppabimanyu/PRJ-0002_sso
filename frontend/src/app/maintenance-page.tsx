import { WrenchIcon } from "lucide-react";

import { StatusPage } from "@/components/shared/status-page";
import { Button } from "@/components/ui/button";

export function MaintenancePage() {
  return (
    <StatusPage
      icon={<WrenchIcon aria-hidden="true" />}
      title="Down for maintenance"
      description="We are making changes and will be back shortly. Nothing is lost; reload this page in a few minutes."
      actions={<Button onClick={() => window.location.reload()}>Try again</Button>}
    />
  );
}

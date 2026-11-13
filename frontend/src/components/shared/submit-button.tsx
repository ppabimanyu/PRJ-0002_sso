import type { ComponentProps, ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

type SubmitButtonProps = Omit<ComponentProps<typeof Button>, "type"> & {
  isSubmitting?: boolean;
  pendingLabel?: ReactNode;
};

export function SubmitButton({
  isSubmitting = false,
  pendingLabel = "Saving",
  children,
  disabled,
  ...props
}: SubmitButtonProps) {
  return (
    <Button
      type="submit"
      disabled={disabled || isSubmitting}
      aria-busy={isSubmitting || undefined}
      {...props}
    >
      {isSubmitting ? (
        <Spinner aria-label={typeof pendingLabel === "string" ? pendingLabel : "Saving"} />
      ) : null}
      {isSubmitting ? pendingLabel : children}
    </Button>
  );
}

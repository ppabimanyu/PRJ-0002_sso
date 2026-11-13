import { Check, Copy } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";

type CopyButtonProps = {
  value: string;
  label?: string;
  className?: string;
};

export function CopyButton({ value, label = "Copy", className }: CopyButtonProps) {
  const [copied, setCopied] = useState(false);
  const timeout = useRef<number | undefined>(undefined);

  useEffect(() => {
    return () => window.clearTimeout(timeout.current);
  }, []);

  async function handleCopy() {
    if (!navigator.clipboard) {
      toast.error("Copying is not available in this browser.");
      return;
    }

    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success("Copied to clipboard");
      window.clearTimeout(timeout.current);
      timeout.current = window.setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Could not copy to the clipboard.");
    }
  }

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      className={className}
      aria-label={copied ? "Copied" : label}
      onClick={() => void handleCopy()}
    >
      {copied ? <Check /> : <Copy />}
      {copied ? "Copied" : label}
    </Button>
  );
}

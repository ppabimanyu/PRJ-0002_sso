import { Search, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebounce } from "@/hooks/use-debounce";
import { cn } from "@/lib/utils";

type SearchInputProps = {
  value: string;
  onValueChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
  className?: string;
};

/**
 * Debounces keystrokes before reporting them, so a list query only reruns once
 * the user pauses. The clear action reports immediately.
 */
export function SearchInput({
  value,
  onValueChange,
  placeholder = "Search",
  debounceMs = 300,
  className,
}: SearchInputProps) {
  const [draft, setDraft] = useState(value);
  const debounced = useDebounce(draft, debounceMs);
  const lastEmitted = useRef(value);

  useEffect(() => {
    if (debounced === lastEmitted.current || debounced === value) {
      return;
    }

    lastEmitted.current = debounced;
    onValueChange(debounced);
  }, [debounced, onValueChange, value]);

  useEffect(() => {
    if (value !== lastEmitted.current) {
      lastEmitted.current = value;
      setDraft(value);
    }
  }, [value]);

  function clear() {
    lastEmitted.current = "";
    setDraft("");
    onValueChange("");
  }

  return (
    <div className={cn("relative w-full max-w-sm", className)}>
      <Search
        className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
        aria-hidden="true"
      />
      <Input
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        placeholder={placeholder}
        aria-label={placeholder}
        className="pl-9 pr-9"
      />
      {draft ? (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="absolute top-1/2 right-1 size-7 -translate-y-1/2"
          aria-label="Clear search"
          onClick={clear}
        >
          <X />
        </Button>
      ) : null}
    </div>
  );
}

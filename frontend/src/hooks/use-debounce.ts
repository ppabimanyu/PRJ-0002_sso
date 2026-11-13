import { useEffect, useState } from "react";

/**
 * Returns `value` after it has stayed unchanged for `delayMs`. Used by inputs
 * that must not write to the URL on every keystroke.
 */
export function useDebounce<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timeout = window.setTimeout(() => setDebounced(value), delayMs);

    return () => window.clearTimeout(timeout);
  }, [value, delayMs]);

  return debounced;
}

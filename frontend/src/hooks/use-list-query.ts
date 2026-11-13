import { useCallback } from "react";
import { useSearchParams } from "react-router-dom";

// Mirrors the backend pagination defaults so a list request built from these
// values is valid without extra clamping.
export const DEFAULT_LIMIT = 20;
export const MAX_LIMIT = 100;
export const LIMIT_OPTIONS = [10, 20, 50, 100] as const;

function readLimit(value: string | null): number {
  const parsed = Number(value);

  if (!Number.isInteger(parsed) || parsed < 1 || parsed > MAX_LIMIT) {
    return DEFAULT_LIMIT;
  }

  return parsed;
}

function readOffset(value: string | null): number {
  const parsed = Number(value);

  return Number.isInteger(parsed) && parsed >= 0 ? parsed : 0;
}

export type ListQuery = {
  limit: number;
  offset: number;
  q: string;
  setLimit: (limit: number) => void;
  setOffset: (offset: number) => void;
  setQuery: (q: string) => void;
};

/**
 * Owns the `limit`, `offset`, and `q` search params for a list route. Pages use
 * this instead of reading `useSearchParams` directly, so every list screen
 * shares one parsing and reset behaviour.
 */
export function useListQuery(): ListQuery {
  const [searchParams, setSearchParams] = useSearchParams();

  const limit = readLimit(searchParams.get("limit"));
  const offset = readOffset(searchParams.get("offset"));
  const q = searchParams.get("q") ?? "";

  const write = useCallback(
    (changes: Record<string, string | null>) => {
      setSearchParams((previous) => {
        const next = new URLSearchParams(previous);

        for (const [key, value] of Object.entries(changes)) {
          if (value === null || value === "") {
            next.delete(key);
          } else {
            next.set(key, value);
          }
        }

        return next;
      });
    },
    [setSearchParams],
  );

  const setLimit = useCallback(
    (next: number) => write({ limit: String(next), offset: null }),
    [write],
  );

  const setOffset = useCallback(
    (next: number) => write({ offset: next === 0 ? null : String(next) }),
    [write],
  );

  const setQuery = useCallback((next: string) => write({ q: next, offset: null }), [write]);

  return { limit, offset, q, setLimit, setOffset, setQuery };
}

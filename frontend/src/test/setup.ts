import "@testing-library/jest-dom/vitest";

import { cleanup } from "@testing-library/react";
import { setupServer } from "msw/node";
import { afterAll, afterEach } from "vitest";

// jsdom's Request requires an absolute URL. Resolve relative paths against the
// document origin so the same-origin health client works under test.
const NativeRequest = globalThis.Request;

class RequestWithBase extends NativeRequest {
  constructor(input: RequestInfo | URL, init?: RequestInit) {
    super(typeof input === "string" ? new URL(input, window.location.origin) : input, init);
  }
}

globalThis.Request = RequestWithBase;

export const server = setupServer();

server.listen({ onUnhandledFrame: "error" });

afterEach(() => {
  cleanup();
  server.resetHandlers();
});

afterAll(() => server.close());

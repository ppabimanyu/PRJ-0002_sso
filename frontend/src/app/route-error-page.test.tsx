import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { RouteErrorPage } from "./route-error-page";

function ThrowingPage(): never {
  throw new Error("The dashboard failed to render");
}

let consoleError: ReturnType<typeof vi.spyOn>;

beforeEach(() => {
  consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
  consoleError.mockRestore();
});

describe("RouteErrorPage", () => {
  it("replaces a crashing route with the app fallback", async () => {
    const router = createMemoryRouter([
      { path: "/", element: <ThrowingPage />, errorElement: <RouteErrorPage /> },
    ]);

    render(<RouterProvider router={router} />);

    expect(await screen.findByText("Something went wrong")).toBeInTheDocument();
    expect(screen.getByText("The dashboard failed to render")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Reload page" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go home" })).toHaveAttribute("href", "/");
  });

  it("shows a not found page for a 404 route response", async () => {
    const router = createMemoryRouter([
      {
        path: "/",
        loader: () => {
          throw new Response(null, { status: 404, statusText: "Not Found" });
        },
        element: <p>hidden</p>,
        errorElement: <RouteErrorPage />,
      },
    ]);

    render(<RouterProvider router={router} />);

    expect(await screen.findByText("Page not found")).toBeInTheDocument();
    expect(screen.getByText("404 Not Found")).toBeInTheDocument();
  });
});

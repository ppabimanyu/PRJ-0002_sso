import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { LoginPage } from "@/features/auth/login-page";
import { renderWithQueryClient } from "@/test/render-with-query-client";
import { server } from "@/test/setup";

const admin = {
  id: "0192f000-0000-7000-8000-000000000001",
  email: "admin@example.com",
  name: "Admin",
  isAdmin: true,
  emailVerified: true,
  createdAt: "2026-11-13T08:00:00Z",
};

async function beginSignIn(email = "admin@example.com") {
  await userEvent.type(await screen.findByLabelText("Email"), email);
  await userEvent.click(screen.getByRole("button", { name: "Continue" }));
}

function renderLogin() {
  return renderWithQueryClient(
    <MemoryRouter initialEntries={["/login"]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/account/profile" element={<div>Account profile</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("LoginPage", () => {
  it("signs in with email and password, then opens the account", async () => {
    server.use(
      http.get("*/api/v1/auth/session", () => HttpResponse.json({ authenticated: false })),
      http.post("*/api/v1/auth/login", async ({ request }) => {
        expect(await request.json()).toEqual({
          email: "admin@example.com",
          password: "correct horse battery staple",
        });
        return HttpResponse.json({ user: admin });
      }),
    );

    renderLogin();

    await beginSignIn();
    await userEvent.type(screen.getByLabelText("Password"), "correct horse battery staple");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Account profile")).toBeInTheDocument();
  });

  it("shows a non-revealing message for bad credentials", async () => {
    server.use(
      http.get("*/api/v1/auth/session", () => HttpResponse.json({ authenticated: false })),
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json(
          {
            type: "about:blank",
            title: "Unauthorized",
            status: 401,
            detail: "Email or password is incorrect.",
          },
          { status: 401, headers: { "Content-Type": "application/problem+json" } },
        ),
      ),
    );

    renderLogin();

    await beginSignIn();
    await userEvent.type(screen.getByLabelText("Password"), "wrong password");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Email or password is incorrect.")).toBeInTheDocument();
  });

  it("explains a temporary lockout", async () => {
    server.use(
      http.get("*/api/v1/auth/session", () => HttpResponse.json({ authenticated: false })),
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json(
          {
            type: "about:blank",
            title: "Too Many Requests",
            status: 429,
            detail: "Too many failed login attempts. Try again later.",
          },
          { status: 429, headers: { "Content-Type": "application/problem+json" } },
        ),
      ),
    );

    renderLogin();

    await beginSignIn();
    await userEvent.type(screen.getByLabelText("Password"), "wrong password");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByText("Too many failed attempts. Try again later.")).toBeInTheDocument();
    });
  });
});

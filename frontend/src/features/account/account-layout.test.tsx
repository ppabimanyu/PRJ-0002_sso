import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AccountLayout } from "@/features/account/account-layout";
import { ProfilePage } from "@/features/account/profile-page";
import { renderWithQueryClient } from "@/test/render-with-query-client";
import { server } from "@/test/setup";

const profile = {
  id: "0192f000-0000-7000-8000-000000000001",
  email: "admin@example.com",
  name: "Admin",
  isAdmin: true,
  emailVerified: true,
  createdAt: "2026-11-13T08:00:00Z",
};

const member = {
  ...profile,
  email: "member@example.com",
  name: "Member",
  isAdmin: false,
};

function renderAccount(user: typeof profile) {
  server.use(
    http.get("*/api/v1/auth/session", () =>
      HttpResponse.json({
        authenticated: true,
        user,
        expiresAt: "2026-11-14T08:00:00Z",
      }),
    ),
    http.post("*/api/v1/auth/logout", () => new HttpResponse(null, { status: 204 })),
  );

  return renderWithQueryClient(
    <MemoryRouter initialEntries={["/account/profile"]}>
      <Routes>
        <Route path="/account" element={<AccountLayout />}>
          <Route path="profile" element={<ProfilePage />} />
        </Route>
        <Route path="/admin" element={<div>Admin console</div>} />
        <Route path="/login" element={<div>Login page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AccountLayout", () => {
  it("shows the profile and an admin link for an Admin", async () => {
    renderAccount(profile);

    expect(await screen.findAllByText("admin@example.com")).not.toHaveLength(0);
    expect(screen.getByText("Verified")).toBeInTheDocument();

    const adminLink = screen.getByRole("link", { name: "Admin" });
    expect(adminLink).toHaveAttribute("href", "/admin");
  });

  it("hides the admin link for a non-Admin", async () => {
    renderAccount(member);

    expect(await screen.findAllByText("member@example.com")).not.toHaveLength(0);
    expect(screen.queryByRole("link", { name: "Admin" })).not.toBeInTheDocument();
    expect(screen.getByText("User")).toBeInTheDocument();
  });

  it("signs out and returns to the login page", async () => {
    renderAccount(profile);

    await screen.findAllByText("admin@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Sign out" }));

    await waitFor(() => {
      expect(screen.getByText("Login page")).toBeInTheDocument();
    });
  });
});

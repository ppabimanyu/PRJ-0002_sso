import { createBrowserRouter, Navigate } from "react-router-dom";

import { AccountLayout } from "@/features/account/account-layout";
import { ProfilePage } from "@/features/account/profile-page";
import { AdminPage } from "@/features/admin/admin-page";
import { LoginPage } from "@/features/auth/login-page";
import { RequireAdmin, RequireAuth } from "@/features/auth/require-auth";
import { HomePage } from "./home-page";
import { Layout } from "./layout";
import { MaintenancePage } from "./maintenance-page";
import { NotFoundPage } from "./not-found-page";
import { RouteErrorPage } from "./route-error-page";
import { RouteFallback } from "./route-fallback";

export const router = createBrowserRouter([
  {
    path: "/maintenance",
    element: <MaintenancePage />,
  },
  {
    path: "/",
    element: <Layout />,
    errorElement: <RouteErrorPage />,
    HydrateFallback: RouteFallback,
    children: [
      { index: true, element: <HomePage /> },
      { path: "login", element: <LoginPage /> },
      {
        element: <RequireAuth />,
        children: [
          {
            path: "account",
            element: <AccountLayout />,
            children: [
              { index: true, element: <Navigate to="/account/profile" replace /> },
              { path: "profile", element: <ProfilePage /> },
            ],
          },
          {
            element: <RequireAdmin />,
            children: [{ path: "admin", element: <AdminPage /> }],
          },
        ],
      },
      ...(import.meta.env.DEV
        ? [
            {
              path: "_gallery",
              lazy: () =>
                import("./gallery-page").then((module) => ({
                  Component: module.GalleryPage,
                })),
            },
          ]
        : []),
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);

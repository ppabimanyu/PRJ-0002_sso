import { createBrowserRouter } from "react-router-dom";

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

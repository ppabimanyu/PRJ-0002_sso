import { Navigate, Outlet, useLocation } from "react-router-dom";

import { RouteFallback } from "@/app/route-fallback";
import { useSession } from "@/features/auth/use-session";

export function RequireAuth() {
  const session = useSession();
  const location = useLocation();

  if (session.isPending) {
    return <RouteFallback />;
  }

  if (!session.data?.authenticated || !session.data.user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <Outlet />;
}

export function RequireAdmin() {
  const session = useSession();

  if (session.isPending) {
    return <RouteFallback />;
  }

  if (!session.data?.user?.isAdmin) {
    return <Navigate to="/account/profile" replace />;
  }

  return <Outlet />;
}

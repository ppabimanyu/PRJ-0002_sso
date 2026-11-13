import { Shield, UserRound } from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";

import { PageHeader } from "@/components/shared/page-header";
import { SignOutButton } from "@/features/auth/sign-out-button";
import { useSession } from "@/features/auth/use-session";
import { cn } from "@/lib/utils";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "flex items-center gap-2 rounded-2xl px-3 py-2 text-sm transition-colors",
    isActive
      ? "bg-muted font-medium text-foreground"
      : "text-muted-foreground hover:bg-muted/60 hover:text-foreground",
  );

export function AccountLayout() {
  const session = useSession();
  const user = session.data?.user;

  return (
    <div className="flex flex-col gap-8">
      <PageHeader
        title="Account"
        description="Your profile and security settings for this Instance."
        actions={<SignOutButton />}
      />

      <div className="grid gap-8 md:grid-cols-[200px_minmax(0,1fr)]">
        <nav aria-label="Account" className="flex flex-col gap-1">
          <NavLink to="/account/profile" className={navLinkClass}>
            <UserRound aria-hidden="true" className="size-4" />
            Profile
          </NavLink>
          {user?.isAdmin ? (
            <NavLink to="/admin" className={navLinkClass}>
              <Shield aria-hidden="true" className="size-4" />
              Admin
            </NavLink>
          ) : null}
        </nav>

        <div className="min-w-0">
          <Outlet />
        </div>
      </div>
    </div>
  );
}

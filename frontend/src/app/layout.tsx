import { Link, NavLink, Outlet } from "react-router-dom";

import { BrandLockup, BrandMark } from "@/brand/brand-mark";
import { appName } from "@/brand/brand";
import { ServiceChart } from "@/brand/service-chart";
import { ThemeToggle } from "@/components/shared/theme-toggle";
import { useSession } from "@/features/auth/use-session";
import { cn } from "@/lib/utils";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "text-sm transition-colors",
    isActive ? "text-foreground" : "text-muted-foreground hover:text-foreground",
  );

export function Layout() {
  const session = useSession();
  const authenticated = session.data?.authenticated === true;

  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <div className="h-[3px] bg-brand" aria-hidden="true" />
      <header className="border-b border-border">
        <div className="mx-auto flex h-14 w-full max-w-[1120px] items-center justify-between px-6">
          <div className="flex items-center gap-8">
            <Link to="/" aria-label={`${appName} home`}>
              <BrandLockup />
            </Link>
            <nav className="flex items-center gap-6">
              <NavLink to="/" end className={navLinkClass}>
                Home
              </NavLink>
              {authenticated ? (
                <NavLink to="/account" className={navLinkClass}>
                  Account
                </NavLink>
              ) : (
                <NavLink to="/login" className={navLinkClass}>
                  Sign in
                </NavLink>
              )}
              {import.meta.env.DEV ? (
                <NavLink to="/_gallery" className={navLinkClass}>
                  Gallery
                </NavLink>
              ) : null}
            </nav>
          </div>
          <ThemeToggle />
        </div>
      </header>
      <main className="mx-auto w-full max-w-[1120px] flex-1 px-6 py-12 sm:py-16">
        <Outlet />
      </main>
      <footer className="border-t border-border">
        <div className="mx-auto flex w-full max-w-[1120px] items-center gap-2.5 px-6 py-8 text-sm text-muted-foreground">
          <BrandMark className="size-4" />
          <span>Go, Postgres, and React, wired together.</span>
          <ServiceChart className="ml-auto w-20 gap-1" />
        </div>
      </footer>
    </div>
  );
}

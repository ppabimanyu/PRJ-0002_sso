import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { useLogout } from "@/features/auth/use-session";

export function SignOutButton() {
  const logout = useLogout();
  const navigate = useNavigate();

  async function handleSignOut() {
    try {
      await logout.mutateAsync();
      navigate("/login", { replace: true });
    } catch {
      toast.error("Could not sign out. Try again.");
    }
  }

  return (
    <Button variant="outline" onClick={() => void handleSignOut()} disabled={logout.isPending}>
      {logout.isPending ? "Signing out" : "Sign out"}
    </Button>
  );
}

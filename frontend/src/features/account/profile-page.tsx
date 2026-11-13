import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { SettingsSection } from "@/components/shared/settings-section";
import { useSession } from "@/features/auth/use-session";

function initials(name: string, email: string) {
  const source = name.trim() || email;
  const parts = source.split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
  }
  return source.slice(0, 2).toUpperCase();
}

function formatDate(value?: string) {
  if (!value) {
    return "Unknown";
  }
  return new Intl.DateTimeFormat("en", { dateStyle: "medium" }).format(new Date(value));
}

export function ProfilePage() {
  const session = useSession();
  const user = session.data?.user;

  if (!user) {
    return null;
  }

  return (
    <div className="flex flex-col gap-6">
      <SettingsSection
        title="Profile"
        description="This identity is shared by every Project that trusts this Instance."
      >
        <div className="flex items-center gap-4">
          <Avatar size="lg">
            {user.pictureUrl ? <AvatarImage src={user.pictureUrl} alt="" /> : null}
            <AvatarFallback>{initials(user.name ?? "", user.email ?? "")}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <p className="truncate font-medium">{user.name}</p>
            <p className="truncate text-sm text-muted-foreground">{user.email}</p>
          </div>
          {user.isAdmin ? <Badge className="ml-auto">Admin</Badge> : null}
        </div>

        <Separator />

        <dl className="grid gap-5 text-sm sm:grid-cols-2">
          <div className="flex flex-col gap-1">
            <dt className="text-muted-foreground">Email</dt>
            <dd className="flex flex-wrap items-center gap-2">
              <span>{user.email}</span>
              <Badge variant={user.emailVerified ? "secondary" : "outline"}>
                {user.emailVerified ? "Verified" : "Unverified"}
              </Badge>
            </dd>
          </div>
          <div className="flex flex-col gap-1">
            <dt className="text-muted-foreground">Instance role</dt>
            <dd>{user.isAdmin ? "Admin" : "User"}</dd>
          </div>
          <div className="flex flex-col gap-1">
            <dt className="text-muted-foreground">Member since</dt>
            <dd>{formatDate(user.createdAt)}</dd>
          </div>
          <div className="flex flex-col gap-1">
            <dt className="text-muted-foreground">Sign-in method</dt>
            <dd>Password</dd>
          </div>
        </dl>
      </SettingsSection>

      <SettingsSection
        title="Security"
        description="Password changes, MFA, and Session management arrive in the next milestone."
      >
        <p className="text-sm leading-relaxed text-muted-foreground">
          Your Session uses the host-only SSO cookie. Signing out revokes it on this device.
        </p>
      </SettingsSection>
    </div>
  );
}

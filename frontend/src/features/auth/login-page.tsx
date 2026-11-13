import { useState, type FormEvent } from "react";
import { Navigate, useLocation } from "react-router-dom";

import { BrandMark } from "@/brand/brand-mark";
import { appName } from "@/brand/brand";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { SubmitButton } from "@/components/shared/submit-button";
import { useLogin, useSession } from "@/features/auth/use-session";
import { ApiError } from "@/lib/problem";

function loginErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.status === 429) {
      return "Too many failed attempts. Try again later.";
    }
    if (error.status === 401) {
      return "Email or password is incorrect.";
    }
    return error.problem?.detail ?? error.message;
  }
  return "Something went wrong. Try again.";
}

export function LoginPage() {
  const session = useSession();
  const login = useLogin();
  const location = useLocation();
  const [step, setStep] = useState<"email" | "password">("email");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  const from = (location.state as { from?: string } | null)?.from ?? "/account/profile";

  if (session.isPending) {
    return (
      <div className="flex justify-center py-24">
        <Spinner aria-label="Checking your session" />
      </div>
    );
  }

  if (session.data?.authenticated) {
    return <Navigate to={from} replace />;
  }

  function handleEmailSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setStep("password");
  }

  async function handlePasswordSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    try {
      await login.mutateAsync({ email, password });
    } catch (submitError) {
      setError(loginErrorMessage(submitError));
    }
  }

  function changeEmail() {
    setError(null);
    setPassword("");
    setStep("email");
  }

  return (
    <div className="mx-auto flex w-full max-w-md flex-col gap-8 py-8 sm:py-16">
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2 text-brand">
          <BrandMark className="size-5" />
          <span className="text-sm font-medium">{appName}</span>
        </div>
        <h1 className="text-3xl font-semibold tracking-tight">Sign in</h1>
        <p className="text-muted-foreground">
          Use the email and password for your identity on this Instance.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{step === "email" ? "Your email" : "Your password"}</CardTitle>
          <CardDescription>
            One account works across every Project that trusts this SSO service.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {step === "email" ? (
            <form className="flex flex-col gap-5" onSubmit={handleEmailSubmit}>
              {error ? (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              ) : null}

              <div className="flex flex-col gap-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  autoComplete="username"
                  autoFocus
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  required
                />
              </div>

              <Button type="submit" className="mt-1 w-full">
                Continue
              </Button>
            </form>
          ) : (
            <form className="flex flex-col gap-5" onSubmit={handlePasswordSubmit}>
              {error ? (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              ) : null}

              <div className="flex items-center justify-between gap-3 rounded-2xl border border-border px-3 py-2 text-sm">
                <span className="truncate">{email}</span>
                <Button type="button" variant="link" className="h-auto p-0" onClick={changeEmail}>
                  Change
                </Button>
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  autoFocus
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                />
              </div>

              <SubmitButton
                className="mt-1 w-full"
                isSubmitting={login.isPending}
                pendingLabel="Signing in"
              >
                Sign in
              </SubmitButton>
            </form>
          )}
        </CardContent>
      </Card>

      <p className="text-sm text-muted-foreground">
        Registration arrives in the next milestone.
      </p>
    </div>
  );
}

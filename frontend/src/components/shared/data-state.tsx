import { Inbox } from "lucide-react";
import type { ReactNode } from "react";

import { Alert, AlertAction, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Spinner } from "@/components/ui/spinner";

type DataStateProps =
  | { status: "loading"; label?: ReactNode }
  | {
      status: "error";
      error: Error;
      title?: ReactNode;
      retryLabel?: ReactNode;
      onRetry?: () => void;
    }
  | {
      status: "empty";
      title?: ReactNode;
      description?: ReactNode;
      action?: ReactNode;
    }
  | { status: "ready"; children: ReactNode };

export function DataState(props: DataStateProps) {
  switch (props.status) {
    case "loading": {
      const label = props.label ?? "Loading";
      return (
        <div className="flex min-h-40 flex-col items-center justify-center gap-3">
          <Spinner className="size-5" aria-label={typeof label === "string" ? label : "Loading"} />
          <span className="text-sm text-muted-foreground">{label}</span>
        </div>
      );
    }
    case "error":
      return (
        <Alert variant="destructive">
          <AlertTitle>{props.title ?? "Something went wrong"}</AlertTitle>
          <AlertDescription>
            {props.error.message || "The request could not be completed."}
          </AlertDescription>
          {props.onRetry ? (
            <AlertAction>
              <Button variant="outline" size="sm" onClick={props.onRetry}>
                {props.retryLabel ?? "Try again"}
              </Button>
            </AlertAction>
          ) : null}
        </Alert>
      );
    case "empty":
      return (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Inbox aria-hidden="true" />
            </EmptyMedia>
            <EmptyTitle>{props.title ?? "No data"}</EmptyTitle>
            {props.description ? <EmptyDescription>{props.description}</EmptyDescription> : null}
          </EmptyHeader>
          {props.action ? <EmptyContent>{props.action}</EmptyContent> : null}
        </Empty>
      );
    case "ready":
      return <>{props.children}</>;
    default: {
      const exhaustive: never = props;
      return exhaustive;
    }
  }
}

export type Problem = {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  [extension: string]: unknown;
};

export class ApiError extends Error {
  readonly status: number;
  readonly problem: Problem | null;

  constructor(status: number, message: string, problem: Problem | null = null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.problem = problem;
  }
}

function parseProblem(value: unknown): Problem | null {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return null;
  }

  const candidate = value as Record<string, unknown>;
  const hasProblemField =
    typeof candidate.title === "string" ||
    typeof candidate.detail === "string" ||
    typeof candidate.status === "number";

  return hasProblemField ? (candidate as Problem) : null;
}

export async function toApiError(response: Response): Promise<ApiError> {
  const isProblemDocument = (response.headers.get("content-type") ?? "").includes(
    "application/problem+json",
  );

  if (isProblemDocument) {
    const body: unknown = await response.json().catch(() => null);
    const problem = parseProblem(body);

    if (problem) {
      const message =
        problem.detail ?? problem.title ?? `Request failed with status ${response.status}`;
      return new ApiError(response.status, message, problem);
    }
  }

  return new ApiError(response.status, `Request failed with status ${response.status}`);
}

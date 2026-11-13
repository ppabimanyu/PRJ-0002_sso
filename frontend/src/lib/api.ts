import createClient, { type Middleware } from "openapi-fetch";

import type { paths } from "@/types/api";
import { toApiError } from "./problem";

const errorMiddleware: Middleware = {
  async onResponse({ response }) {
    if (!response.ok) {
      throw await toApiError(response);
    }

    return response;
  },
};

// ADR 0005 keeps the auth UI and API on one origin, and generated paths already
// include /api/v1. The empty base URL makes every request same-origin.
export const apiClient = createClient<paths>({ baseUrl: "" });

apiClient.use(errorMiddleware);

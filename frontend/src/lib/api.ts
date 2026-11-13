import createClient, { type Middleware } from "openapi-fetch";

import type { paths } from "@/types/api";
import { apiBaseUrl } from "./config";
import { toApiError } from "./problem";

const errorMiddleware: Middleware = {
  async onResponse({ response }) {
    if (!response.ok) {
      throw await toApiError(response);
    }

    return response;
  },
};

export const apiClient = createClient<paths>({ baseUrl: apiBaseUrl });

export const healthClient = createClient<paths>({ baseUrl: "" });

apiClient.use(errorMiddleware);
healthClient.use(errorMiddleware);

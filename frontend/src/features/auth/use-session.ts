import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "@/lib/api";
import type { components } from "@/types/api";

export type SessionUser = components["schemas"]["httphandler.UserResponse"];
export type SessionState = components["schemas"]["httphandler.SessionResponse"];

export const sessionQueryKey = ["auth", "session"] as const;

export function useSession() {
  return useQuery({
    queryKey: sessionQueryKey,
    queryFn: async (): Promise<SessionState> => {
      const { data } = await apiClient.GET("/api/v1/auth/session");
      return data ?? { authenticated: false };
    },
    retry: false,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { email: string; password: string }): Promise<SessionUser> => {
      const { data } = await apiClient.POST("/api/v1/auth/login", { body: input });
      if (!data?.user) {
        throw new Error("Login response did not include a user");
      }
      return data.user;
    },
    onSuccess: (user) => {
      queryClient.setQueryData<SessionState>(sessionQueryKey, { authenticated: true, user });
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await apiClient.POST("/api/v1/auth/logout");
    },
    onSuccess: () => {
      queryClient.setQueryData<SessionState>(sessionQueryKey, { authenticated: false });
    },
  });
}

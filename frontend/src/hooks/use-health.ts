import { useQuery } from "@tanstack/react-query";

import { healthClient } from "@/lib/api";

export function useHealth() {
  return useQuery({
    queryKey: ["health", "ready"],
    queryFn: async () => {
      const { data } = await healthClient.GET("/readyz");

      if (!data) {
        throw new Error("Readiness check returned an empty response");
      }

      return data;
    },
    refetchInterval: 30_000,
    retry: 1,
  });
}

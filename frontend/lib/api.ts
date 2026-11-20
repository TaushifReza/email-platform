import type { components } from "./types/openapi";

export type HealthResponse = components["schemas"]["HealthResponse"];
export type ReadyResponse = components["schemas"]["ReadyResponse"];

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: { accept: "application/json" },
  });

  if (!response.ok) {
    throw new Error(`Request to ${path} failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export function getHealth(): Promise<HealthResponse> {
  return getJSON<HealthResponse>("/api/healthz");
}

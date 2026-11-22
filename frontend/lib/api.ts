import type { components } from "./types/openapi";

export type HealthResponse = components["schemas"]["HealthResponse"];
export type ReadyResponse = components["schemas"]["ReadyResponse"];
export type Problem = components["schemas"]["Problem"];

export class ApiError extends Error {
  status: number;
  code: string;
  requestId: string;
  constructor(problem: Problem, status: number) {
    super(problem.detail ?? problem.title);
    this.name = "ApiError";
    this.status = status;
    this.code = problem.code;
    this.requestId = problem.request_id;
  }
}

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: { accept: "application/json" },
  });

  if (!response.ok) {
    let problem: Problem | null = null;
    try {
      problem = (await response.json()) as Problem;
    } catch {
      problem = null;
    }
    if (problem && problem.code) {
      throw new ApiError(problem, response.status);
    }
    throw new Error(`Request to ${path} failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export function getHealth(): Promise<HealthResponse> {
  return getJSON<HealthResponse>("/api/healthz");
}
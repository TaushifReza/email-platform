import { getHealth } from "./api";

export async function checkHealth(): Promise<{ status: string }> {
  return getHealth();
}

import type { LoginResponse } from "@/lib/api";

export type AuthDestination = "/mfa/enroll" | "/mfa" | "/";

export function nextRouteAfterLogin(response: LoginResponse): AuthDestination {
  if (response.mfa_enrollment_required) return "/mfa/enroll";
  if (response.mfa_required) return "/mfa";
  return "/";
}

export function parseInvitationToken(fragment: string): string | null {
  const cleaned = fragment.replace(/^#/, "");
  const params = new URLSearchParams(cleaned);
  const token = params.get("token");
  if (!token || token.trim() === "") return null;
  return token.trim();
}

export function canManageUsers(role: string): boolean {
  return role === "owner" || role === "admin";
}

export function canEditPolicy(role: string): boolean {
  return role === "owner";
}
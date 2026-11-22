import type { components } from "./types/openapi";

export type HealthResponse = components["schemas"]["HealthResponse"];
export type ReadyResponse = components["schemas"]["ReadyResponse"];
export type Problem = components["schemas"]["Problem"];
export type LoginResponse = components["schemas"]["LoginResponse"];
export type SessionResponse = components["schemas"]["SessionResponse"];
export type Organization = components["schemas"]["Organization"];
export type User = components["schemas"]["User"];
export type UserList = components["schemas"]["UserList"];
export type MFAEnrollResponse = components["schemas"]["MFAEnrollResponse"];
export type MFARecoveryResponse = components["schemas"]["MFARecoveryResponse"];

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

function csrfToken(): string {
  return "frontend";
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method ?? "GET").toUpperCase();
  const headers = new Headers(init.headers);
  headers.set("accept", "application/json");
  if (method !== "GET" && method !== "HEAD") {
    headers.set("X-CSRF-Token", csrfToken());
    if (init.body && !headers.has("content-type")) {
      headers.set("content-type", "application/json");
    }
  }
  const response = await fetch(path, {
    ...init,
    method,
    headers,
    credentials: "include",
  });

  if (response.status === 204) {
    return undefined as T;
  }
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
  return request<HealthResponse>("/api/healthz");
}

export function login(email: string, password: string): Promise<LoginResponse> {
  return request<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export function logout(): Promise<void> {
  return request<void>("/api/auth/logout", { method: "POST" });
}

export function getSession(): Promise<SessionResponse> {
  return request<SessionResponse>("/api/auth/session");
}

export function enrollMFA(): Promise<MFAEnrollResponse> {
  return request<MFAEnrollResponse>("/api/auth/mfa/enroll", { method: "POST" });
}

export function confirmMFA(code: string): Promise<MFARecoveryResponse> {
  return request<MFARecoveryResponse>("/api/auth/mfa/enroll/confirm", {
    method: "POST",
    body: JSON.stringify({ code }),
  });
}

export function verifyMFA(code: string): Promise<void> {
  return request<void>("/api/auth/mfa/verify", {
    method: "POST",
    body: JSON.stringify({ code }),
  });
}

export function stepUp(password: string, code: string): Promise<void> {
  return request<void>("/api/auth/step-up", {
    method: "POST",
    body: JSON.stringify({ password, code }),
  });
}

export function acceptInvitation(token: string, password: string): Promise<void> {
  return request<void>("/api/auth/invitations/accept", {
    method: "POST",
    body: JSON.stringify({ token, password }),
  });
}

export function getOrganization(): Promise<Organization> {
  return request<Organization>("/api/organization");
}

export function updateOrganizationProfile(payload: {
  name?: string;
  defaultSenderName?: string;
  defaultReplyTo?: string;
  timeZone?: string;
}): Promise<Organization> {
  return request<Organization>("/api/organization", {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export function updateOrganizationPolicy(payload: {
  requireApproval?: boolean;
  retentionDays?: number;
}): Promise<Organization> {
  return request<Organization>("/api/organization/policy", {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export function listUsers(): Promise<UserList> {
  return request<UserList>("/api/users");
}

export function inviteUser(email: string, role: string): Promise<void> {
  return request<void>("/api/users/invitations", {
    method: "POST",
    body: JSON.stringify({ email, role }),
  });
}

export function updateUser(
  userId: string,
  payload: { role?: string; deactivate?: boolean },
): Promise<void> {
  return request<void>(`/api/users/${encodeURIComponent(userId)}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}
export type OperatorSession = components["schemas"]["OperatorSession"];
export type OrganizationList = components["schemas"]["OrganizationList"];
export type AuditEntryList = components["schemas"]["AuditEntryList"];

export function getOperatorSession(): Promise<OperatorSession> {
  return request<OperatorSession>("/api/operator/session");
}

export function listOperatorOrganizations(limit = 50, offset = 0): Promise<OrganizationList> {
  return request<OrganizationList>(`/api/operator/organizations?limit=${limit}&offset=${offset}`);
}

export function createOperatorOrganization(payload: {
  name: string;
  defaultSenderName?: string;
  defaultReplyTo?: string;
  timeZone?: string;
}): Promise<Organization> {
  return request<Organization>("/api/operator/organizations", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function suspendOrganization(organizationId: string): Promise<void> {
  return request<void>(`/api/operator/organizations/${encodeURIComponent(organizationId)}/suspend`, {
    method: "POST",
  });
}

export function reinstateOrganization(organizationId: string): Promise<void> {
  return request<void>(`/api/operator/organizations/${encodeURIComponent(organizationId)}/reinstate`, {
    method: "POST",
  });
}

export function updateOrganizationLimits(
  organizationId: string,
  payload: {
    storageQuotaBytes: number;
    capsEnabled: boolean;
    dailyEmailCap?: number;
    campaignRecipientCap?: number;
  },
): Promise<Organization> {
  return request<Organization>(`/api/operator/organizations/${encodeURIComponent(organizationId)}/limits`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export function listOperatorAudit(limit = 50, offset = 0): Promise<AuditEntryList> {
  return request<AuditEntryList>(`/api/operator/audit?limit=${limit}&offset=${offset}`);
}

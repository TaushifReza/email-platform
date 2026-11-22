import { afterEach, describe, expect, it, vi } from "vitest";
import { authService } from "@/features/auth/auth-service";

function stubFetch(responses: Array<{ status: number; body?: unknown }>) {
  const fetchMock = vi.fn();
  for (const response of responses) {
    fetchMock.mockResolvedValueOnce({
      ok: response.status >= 200 && response.status < 300,
      status: response.status,
      json: async () => response.body,
    });
  }
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("auth service", () => {
  it("maps login, session and logout calls", async () => {
    const fetchMock = stubFetch([
      { status: 200, body: { mfa_required: false, mfa_enrollment_required: false } },
      { status: 200, body: { user_id: "u1", organization_id: "o1", role: "owner", mfa_verified: true } },
      { status: 204 },
    ]);

    await authService.login("a@b.com", "secret");
    await authService.session();
    await authService.logout();

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/auth/login",
      "/api/auth/session",
      "/api/auth/logout",
    ]);
  });

  it("maps MFA and invitation calls", async () => {
    const fetchMock = stubFetch([
      { status: 200, body: { secret: "s", otpauth_uri: "uri" } },
      { status: 200, body: { recovery_codes: ["r1"] } },
      { status: 204 },
      { status: 204 },
      { status: 204 },
    ]);

    await authService.enroll();
    await authService.confirmEnrollment("123456");
    await authService.verify("123456");
    await authService.stepUp("secret", "123456");
    await authService.acceptInvitation("tok", "secret");

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/auth/mfa/enroll",
      "/api/auth/mfa/enroll/confirm",
      "/api/auth/mfa/verify",
      "/api/auth/step-up",
      "/api/auth/invitations/accept",
    ]);
  });
});
import { describe, expect, it } from "vitest";
import {
  canEditPolicy,
  canManageUsers,
  nextRouteAfterLogin,
  parseInvitationToken,
} from "@/features/auth/auth-flow";

describe("auth flow", () => {
  it("routes first-login enrollment before MFA verification", () => {
    expect(nextRouteAfterLogin({ mfa_required: true, mfa_enrollment_required: true })).toBe("/mfa/enroll");
    expect(nextRouteAfterLogin({ mfa_required: true, mfa_enrollment_required: false })).toBe("/mfa");
    expect(nextRouteAfterLogin({ mfa_required: false, mfa_enrollment_required: false })).toBe("/");
  });

  it("parses invitation token from the URL fragment", () => {
    expect(parseInvitationToken("#token=abc-123")).toBe("abc-123");
    expect(parseInvitationToken("token=abc-123")).toBe("abc-123");
    expect(parseInvitationToken("#other=1")).toBeNull();
  });

  it("gates management UI by role", () => {
    expect(canManageUsers("owner")).toBe(true);
    expect(canManageUsers("admin")).toBe(true);
    expect(canManageUsers("viewer")).toBe(false);
    expect(canEditPolicy("owner")).toBe(true);
    expect(canEditPolicy("admin")).toBe(false);
  });
});
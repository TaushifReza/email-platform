import { afterEach, describe, expect, it, vi } from "vitest";
import { organizationService } from "@/features/organization/organization-service";

const org = {
  id: "o1",
  name: "Acme",
  defaultSenderName: "Acme",
  defaultReplyTo: "no-reply@example.com",
  timeZone: "UTC",
  status: "active",
  requireApproval: false,
  retentionDays: 365,
  storageQuotaBytes: 1000,
  capsEnabled: false,
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("organization service", () => {
  it("sends profile and policy fields unchanged", async () => {
    const fetchMock = vi.fn();
    fetchMock.mockResolvedValueOnce({ ok: true, status: 200, json: async () => org });
    fetchMock.mockResolvedValueOnce({ ok: true, status: 200, json: async () => org });
    fetchMock.mockResolvedValueOnce({ ok: true, status: 200, json: async () => org });
    vi.stubGlobal("fetch", fetchMock);

    await organizationService.get();
    await organizationService.updateProfile({
      name: "Acme",
      defaultSenderName: "Acme",
      defaultReplyTo: "no-reply@example.com",
      timeZone: "UTC",
    });
    await organizationService.updatePolicy({ requireApproval: true, retentionDays: 30 });

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/organization",
      "/api/organization",
      "/api/organization/policy",
    ]);
    expect(fetchMock.mock.calls[1][1].method).toBe("PATCH");
    expect(fetchMock.mock.calls[2][1].body).toBe('{"requireApproval":true,"retentionDays":30}');
  });
});
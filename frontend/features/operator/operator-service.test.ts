import { afterEach, describe, expect, it, vi } from "vitest";
import { operatorService } from "@/features/operator/operator-service";

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

describe("operator service", () => {
  it("calls the operator session and list endpoints", async () => {
    const fetchMock = stubFetch([
      { status: 200, body: { id: "op-1", email: "a@b.com", role: "platform_admin" } },
      { status: 200, body: { data: [], meta: {} } },
      { status: 200, body: { data: [], meta: {} } },
    ]);

    await operatorService.session();
    await operatorService.listOrganizations(20, 40);
    await operatorService.listAudit();

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/operator/session",
      "/api/operator/organizations?limit=20&offset=40",
      "/api/operator/audit?limit=50&offset=0",
    ]);
  });

  it("sends CSRF header and encoded paths on operator mutations", async () => {
    const fetchMock = stubFetch([
      { status: 201, body: { id: "o1", name: "Acme" } },
      { status: 204 },
      { status: 204 },
      { status: 200, body: { id: "o1" } },
    ]);

    await operatorService.createOrganization({ name: "Acme" });
    await operatorService.suspend("o 1");
    await operatorService.reinstate("o 1");
    await operatorService.updateLimits("o 1", { storageQuotaBytes: 100, capsEnabled: false });

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/operator/organizations",
      "/api/operator/organizations/o%201/suspend",
      "/api/operator/organizations/o%201/reinstate",
      "/api/operator/organizations/o%201/limits",
    ]);
    for (const [, init] of fetchMock.mock.calls) {
      const headers = new Headers(init.headers as HeadersInit);
      expect(headers.get("X-CSRF-Token")).toBeTruthy();
    }
    expect(fetchMock.mock.calls[3][1].body).toBe('{"storageQuotaBytes":100,"capsEnabled":false}');
  });
});

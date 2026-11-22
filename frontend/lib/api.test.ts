import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, login, logout } from "@/lib/api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("api request wrapper", () => {
  it("sends credentials and CSRF header on mutations", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ mfa_required: false, mfa_enrollment_required: false }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await login("a@b.com", "secret");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const headers = new Headers(init.headers);
    expect(init.credentials).toBe("include");
    expect(headers.get("X-CSRF-Token")).toBeTruthy();
    expect(headers.get("content-type")).toBe("application/json");
  });

  it("handles 204 responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 204 });
    vi.stubGlobal("fetch", fetchMock);
    await expect(logout()).resolves.toBeUndefined();
  });

  it("maps problem responses to ApiError", async () => {
    const problem = {
      type: "about:blank",
      title: "Invalid credentials",
      status: 401,
      code: "invalid_credentials",
      request_id: "req-1",
    };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => problem,
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(login("a@b.com", "bad")).rejects.toMatchObject({
      name: "ApiError",
      status: 401,
      code: "invalid_credentials",
      requestId: "req-1",
    });
    expect(ApiError.name).toBe("ApiError");
  });
});
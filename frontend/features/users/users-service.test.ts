import { afterEach, describe, expect, it, vi } from "vitest";
import { usersService } from "@/features/users/users-service";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("users service", () => {
  it("lists, invites and updates users", async () => {
    const fetchMock = vi.fn();
    fetchMock.mockResolvedValueOnce({ ok: true, status: 200, json: async () => ({ data: [], meta: {} }) });
    fetchMock.mockResolvedValueOnce({ ok: true, status: 204 });
    fetchMock.mockResolvedValueOnce({ ok: true, status: 204 });
    fetchMock.mockResolvedValueOnce({ ok: true, status: 204 });
    vi.stubGlobal("fetch", fetchMock);

    await usersService.list();
    await usersService.invite("a@b.com", "viewer");
    await usersService.update("u1", { role: "admin" });
    await usersService.update("u1", { deactivate: true });

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "/api/users",
      "/api/users/invitations",
      "/api/users/u1",
      "/api/users/u1",
    ]);
    expect(fetchMock.mock.calls[1][1].body).toBe('{"email":"a@b.com","role":"viewer"}');
    expect(fetchMock.mock.calls[2][1].body).toBe('{"role":"admin"}');
    expect(fetchMock.mock.calls[3][1].body).toBe('{"deactivate":true}');
  });
});
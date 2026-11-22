"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { authService } from "@/features/auth/auth-service";
import type { SessionResponse } from "@/lib/api";

type SessionState =
  | { status: "loading" }
  | { status: "authenticated"; session: SessionResponse }
  | { status: "unauthenticated" }
  | { status: "mfa-required" };

const SessionContext = createContext<SessionState>({ status: "loading" });

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<SessionState>({ status: "loading" });
  const router = useRouter();

  useEffect(() => {
    authService
      .session()
      .then((session) => setState({ status: "authenticated", session }))
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.code === "mfa_required") {
          setState({ status: "mfa-required" });
          router.replace("/mfa");
          return;
        }
        setState({ status: "unauthenticated" });
        router.replace("/login");
      });
  }, [router]);

  return <SessionContext.Provider value={state}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionState {
  return useContext(SessionContext);
}
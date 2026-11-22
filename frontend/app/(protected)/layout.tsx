"use client";

import Link from "next/link";
import { ApiError } from "@/lib/api";
import { authService } from "@/features/auth/auth-service";
import { SessionProvider, useSession } from "@/features/auth/session-context";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";

function Shell({ children }: { children: React.ReactNode }) {
  const session = useSession();
  const router = useRouter();

  if (session.status !== "authenticated") {
    return <p className="p-8 text-sm">Loading session...</p>;
  }

  async function signOut() {
    try {
      await authService.logout();
    } catch (e: unknown) {
      if (!(e instanceof ApiError)) throw e;
    }
    router.replace("/login");
  }

  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b px-6 py-4">
        <nav className="flex gap-4 text-sm">
          <Link href="/">Dashboard</Link>
          <Link href="/users">Users</Link>
          <Link href="/settings/organization">Organization settings</Link>
        </nav>
        <div className="flex items-center gap-4 text-sm">
          <span>{session.session.role}</span>
          <Button variant="outline" onClick={signOut}>
            Sign out
          </Button>
        </div>
      </header>
      <main className="mx-auto max-w-5xl p-6">{children}</main>
    </div>
  );
}

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <Shell>{children}</Shell>
    </SessionProvider>
  );
}
"use client";

import { useSession } from "@/features/auth/session-context";
import { UserList } from "@/features/users/components/user-list";

export default function UsersPage() {
  const session = useSession();
  const role = session.status === "authenticated" ? session.session.role : "";
  return <UserList role={role} />;
}
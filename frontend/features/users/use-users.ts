"use client";

import { useCallback, useState } from "react";
import type { User } from "@/lib/api";
import { usersService } from "@/features/users/users-service";

export function useUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await usersService.list();
      setUsers(result.data ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to load users");
    } finally {
      setLoading(false);
    }
  }, []);

  return { users, loading, error, reload };
}
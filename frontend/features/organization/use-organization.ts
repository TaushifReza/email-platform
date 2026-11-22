"use client";

import { useEffect, useState } from "react";
import type { Organization } from "@/lib/api";
import { organizationService } from "@/features/organization/organization-service";

export function useOrganization() {
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    organizationService
      .get()
      .then(setOrganization)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : "Failed to load"))
      .finally(() => setLoading(false));
  }, []);

  return { organization, loading, error, setOrganization };
}
"use client";

import { useSession } from "@/features/auth/session-context";
import { OrganizationSettingsForm } from "@/features/organization/components/organization-settings-form";

export default function OrganizationSettingsPage() {
  const session = useSession();
  const role = session.status === "authenticated" ? session.session.role : "";
  return <OrganizationSettingsForm role={role} />;
}
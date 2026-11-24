"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { AdminConsole } from "@/features/operator/components/admin-console";
import { getOperatorSession } from "@/lib/api";

export default function AdminPage() {
  const router = useRouter();
  useEffect(() => {
    getOperatorSession().catch(() => router.replace("/operator/login"));
  }, [router]);
  return (
    <div className="mx-auto max-w-6xl p-6">
      <AdminConsole />
    </div>
  );
}

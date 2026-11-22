import Link from "next/link";

export default function DashboardPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="text-sm">Use the navigation above to manage users and organization settings.</p>
      <div className="flex gap-4 text-sm">
        <Link href="/users" className="underline">Users</Link>
        <Link href="/settings/organization" className="underline">Organization settings</Link>
      </div>
    </div>
  );
}
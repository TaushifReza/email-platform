"use client";

import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import { canEditPolicy } from "@/features/auth/auth-flow";
import { organizationService } from "@/features/organization/organization-service";
import { useOrganization } from "@/features/organization/use-organization";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";

export function OrganizationSettingsForm({ role }: { role: string }) {
  const { organization, setOrganization } = useOrganization();
  const [name, setName] = useState("");
  const [senderName, setSenderName] = useState("");
  const [replyTo, setReplyTo] = useState("");
  const [timeZone, setTimeZone] = useState("");
  const [requireApproval, setRequireApproval] = useState(false);
  const [retentionDays, setRetentionDays] = useState(365);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!organization) return;
    setName(organization.name ?? "");
    setSenderName(organization.defaultSenderName ?? "");
    setReplyTo(organization.defaultReplyTo ?? "");
    setTimeZone(organization.timeZone ?? "");
    setRequireApproval(Boolean(organization.requireApproval));
    setRetentionDays(organization.retentionDays ?? 365);
  }, [organization]);

  const canPolicy = canEditPolicy(role);

  async function saveProfile(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSaving(true);
    try {
      const updated = await organizationService.updateProfile({
        name,
        defaultSenderName: senderName,
        defaultReplyTo: replyTo,
        timeZone,
      });
      setOrganization(updated);
    } catch (e: unknown) {
      setError(e instanceof ApiError ? e.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  async function savePolicy(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSaving(true);
    try {
      const updated = await organizationService.updatePolicy({
        requireApproval,
        retentionDays,
      });
      setOrganization(updated);
    } catch (e: unknown) {
      setError(e instanceof ApiError ? e.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  if (!organization) return <p className="text-sm">Loading organization...</p>;

  return (
    <div className="space-y-6">
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <Card>
        <CardHeader>
          <CardTitle>Organization profile</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={saveProfile}>
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="senderName">Default sender name</Label>
              <Input id="senderName" value={senderName} onChange={(e) => setSenderName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="replyTo">Default reply-to</Label>
              <Input id="replyTo" value={replyTo} onChange={(e) => setReplyTo(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="timeZone">Time zone</Label>
              <Input id="timeZone" value={timeZone} onChange={(e) => setTimeZone(e.target.value)} />
            </div>
            <Button type="submit" disabled={saving}>
              {saving ? "Saving..." : "Save profile"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Organization policy</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={savePolicy}>
            <div className="space-y-2">
              <Label htmlFor="requireApproval">Require campaign approval</Label>
              <Select
                id="requireApproval"
                value={requireApproval ? "true" : "false"}
                onChange={(e) => setRequireApproval(e.target.value === "true")}
                disabled={!canPolicy}
              >
                <option value="false">Off</option>
                <option value="true">On</option>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="retentionDays">Retention days</Label>
              <Input
                id="retentionDays"
                type="number"
                value={retentionDays}
                onChange={(e) => setRetentionDays(Number(e.target.value))}
                disabled={!canPolicy}
              />
            </div>
            <Button type="submit" disabled={saving || !canPolicy}>
              {saving ? "Saving..." : "Save policy"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
"use client";

import { useCallback, useEffect, useState } from "react";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TD, TH, THead, TR, Table, TBody } from "@/components/ui/table";
import { ApiError } from "@/lib/api";
import {
  operatorService,
  type OperatorAuditEntry,
  type OperatorOrganization,
} from "@/features/operator/operator-service";

const gb = 1024 * 1024 * 1024;

function message(error: unknown): string {
  return error instanceof ApiError ? error.message : "Request failed";
}

export function AdminConsole() {
  const [session, setSession] = useState<string | null>(null);
  const [orgs, setOrgs] = useState<OperatorOrganization[]>([]);
  const [audit, setAudit] = useState<OperatorAuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [name, setName] = useState("");
  const [sender, setSender] = useState("");
  const [replyTo, setReplyTo] = useState("");
  const [timeZone, setTimeZone] = useState("UTC");

  const refresh = useCallback(async () => {
    const [orgResult, auditResult] = await Promise.all([
      operatorService.listOrganizations(),
      operatorService.listAudit(),
    ]);
    setOrgs(orgResult.data ?? []);
    setAudit(auditResult.data ?? []);
  }, []);

  useEffect(() => {
    let active = true;
    operatorService
      .session()
      .then((s) => {
        if (active) setSession(s.email);
        return refresh();
      })
      .catch((e: unknown) => {
        if (active) setError(message(e));
      });
    return () => {
      active = false;
    };
  }, [refresh]);

  async function run(action: () => Promise<unknown>) {
    setBusy(true);
    setError(null);
    try {
      await action();
      await refresh();
    } catch (e: unknown) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }

  function createOrganization(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void run(async () => {
      await operatorService.createOrganization({
        name,
        defaultSenderName: sender,
        defaultReplyTo: replyTo,
        timeZone,
      });
      setName("");
      setSender("");
      setReplyTo("");
    });
  }

  function changeState(org: OperatorOrganization, suspend: boolean) {
    const id = org.id;
    if (!id) return;
    void run(() => (suspend ? operatorService.suspend(id) : operatorService.reinstate(id)));
  }

  function changeQuota(org: OperatorOrganization) {
    const id = org.id;
    if (!id) return;
    const current = Math.round((org.storageQuotaBytes ?? 0) / gb);
    const value = window.prompt("Storage quota in GB", String(current));
    if (value === null) return;
    const quota = Number(value);
    if (!Number.isInteger(quota) || quota <= 0) {
      setError("Quota must be a positive whole number of GB");
      return;
    }
    void run(() => operatorService.updateLimits(id, {
      storageQuotaBytes: quota * gb,
      capsEnabled: Boolean(org.capsEnabled),
      dailyEmailCap: org.dailyEmailCap ?? undefined,
      campaignRecipientCap: org.campaignRecipientCap ?? undefined,
    }));
  }

  if (error && !session) {
    return (
      <div className="space-y-4">
        <Alert>{error}</Alert>
        <p className="text-sm">
          Open the console on the operator host (local: <code>admin.localhost:8081</code>). The tenant host has no
          operator identity.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-2xl font-semibold">Operator console</h1>
        <p className="text-sm text-muted-foreground">Signed in as {session ?? "..."}</p>
      </header>

      {error ? <Alert>{error}</Alert> : null}

      <Card>
        <CardHeader>
          <CardTitle>Create organization</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4 md:grid-cols-2" onSubmit={createOrganization}>
            <div className="space-y-1">
              <Label htmlFor="org-name">Name</Label>
              <Input id="org-name" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="space-y-1">
              <Label htmlFor="org-sender">Default sender name</Label>
              <Input id="org-sender" value={sender} onChange={(e) => setSender(e.target.value)} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="org-reply">Default reply-to</Label>
              <Input id="org-reply" value={replyTo} onChange={(e) => setReplyTo(e.target.value)} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="org-tz">Time zone</Label>
              <Input id="org-tz" value={timeZone} onChange={(e) => setTimeZone(e.target.value)} />
            </div>
            <div>
              <Button type="submit" disabled={busy}>Create</Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Organizations</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto">
          <Table>
            <THead>
              <TR>
                <TH>Name</TH>
                <TH>Status</TH>
                <TH>Quota (GB)</TH>
                <TH>Caps</TH>
                <TH>Actions</TH>
              </TR>
            </THead>
            <TBody>
              {orgs.map((org) => (
                <TR key={org.id}>
                  <TD>{org.name}</TD>
                  <TD>{org.status}</TD>
                  <TD>{Math.round((org.storageQuotaBytes ?? 0) / gb)}</TD>
                  <TD>{org.capsEnabled ? "on" : "off"}</TD>
                  <TD className="flex gap-2">
                    <Button variant="outline" disabled={busy} onClick={() => changeState(org, org.status !== "suspended")}>
                      {org.status === "suspended" ? "Reinstate" : "Suspend"}
                    </Button>
                    <Button variant="outline" disabled={busy} onClick={() => changeQuota(org)}>
                      Set quota
                    </Button>
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Operator audit</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto">
          <Table>
            <THead>
              <TR>
                <TH>When (UTC)</TH>
                <TH>Action</TH>
                <TH>Actor</TH>
                <TH>Target</TH>
                <TH>Outcome</TH>
              </TR>
            </THead>
            <TBody>
              {audit.map((entry) => (
                <TR key={entry.id}>
                  <TD>{entry.createdAt ? new Date(entry.createdAt).toISOString() : ""}</TD>
                  <TD>{entry.action}</TD>
                  <TD>{entry.actorId}</TD>
                  <TD>{entry.targetId}</TD>
                  <TD>{entry.outcome}</TD>
                </TR>
              ))}
            </TBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}

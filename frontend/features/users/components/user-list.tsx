"use client";

import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import { canManageUsers } from "@/features/auth/auth-flow";
import { usersService } from "@/features/users/users-service";
import { useUsers } from "@/features/users/use-users";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/table";
import { Alert } from "@/components/ui/alert";
import { StepUpDialog } from "@/features/auth/components/step-up-dialog";

const ROLES = ["owner", "admin", "campaign_manager", "reviewer", "viewer"];

export function UserList({ role }: { role: string }) {
  const { users, loading, error, reload } = useUsers();
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("viewer");
  const [message, setMessage] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<
    | { type: "role"; userId: string; role: string }
    | { type: "deactivate"; userId: string }
    | null
  >(null);

  useEffect(() => {
    reload();
  }, [reload]);

  if (!canManageUsers(role)) {
    return <Alert>You do not have permission to manage users.</Alert>;
  }

  async function invite(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage(null);
    try {
      await usersService.invite(inviteEmail, inviteRole);
      setInviteEmail("");
      await reload();
      setMessage("Invitation sent");
    } catch (e: unknown) {
      setMessage(e instanceof ApiError ? e.message : "Invite failed");
    }
  }

  async function runPending() {
    if (!pendingAction) return;
    try {
      if (pendingAction.type === "role") {
        await usersService.update(pendingAction.userId, { role: pendingAction.role });
      } else {
        await usersService.update(pendingAction.userId, { deactivate: true });
      }
      setPendingAction(null);
      await reload();
    } catch (e: unknown) {
      setMessage(e instanceof ApiError ? e.message : "Action failed");
      setPendingAction(null);
    }
  }

  return (
    <div className="space-y-6">
      {message ? <p className="text-sm">{message}</p> : null}
      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      <Card>
        <CardHeader>
          <CardTitle>Invite user</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="flex flex-wrap items-end gap-3" onSubmit={invite}>
            <div className="space-y-2">
              <Label htmlFor="inviteEmail">Email</Label>
              <Input id="inviteEmail" type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="inviteRole">Role</Label>
              <Select id="inviteRole" value={inviteRole} onChange={(e) => setInviteRole(e.target.value)}>
                {ROLES.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </Select>
            </div>
            <Button type="submit">Invite</Button>
          </form>
        </CardContent>
      </Card>

      {pendingAction ? (
        <Card>
          <CardHeader>
            <CardTitle>Step-up required</CardTitle>
          </CardHeader>
          <CardContent>
            <StepUpDialog onVerified={runPending} />
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Users</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? <p className="text-sm">Loading users...</p> : null}
          <Table>
            <THead>
              <TR>
                <TH>Email</TH>
                <TH>Role</TH>
                <TH>Status</TH>
                <TH>Actions</TH>
              </TR>
            </THead>
            <TBody>
              {users.map((user) => (
                <TR key={user.id}>
                  <TD>{user.email}</TD>
                  <TD>{user.role}</TD>
                  <TD>{user.status}</TD>
                  <TD className="space-x-2">
                    <Select
                      value={user.role ?? "viewer"}
                      onChange={(e) =>
                        setPendingAction({ type: "role", userId: user.id ?? "", role: e.target.value })
                      }
                    >
                      {ROLES.map((r) => (
                        <option key={r} value={r}>
                          {r}
                        </option>
                      ))}
                    </Select>
                    <Button
                      variant="outline"
                      onClick={() => setPendingAction({ type: "deactivate", userId: user.id ?? "" })}
                    >
                      Deactivate
                    </Button>
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
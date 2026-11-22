"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { authService } from "@/features/auth/auth-service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function MfaEnroll() {
  const router = useRouter();
  const [secret, setSecret] = useState<string | null>(null);
  const [otpauthUri, setOtpauthUri] = useState<string | null>(null);
  const [code, setCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    authService
      .enroll()
      .then((result) => {
        setSecret(result.secret);
        setOtpauthUri(result.otpauth_uri);
      })
      .catch((e: unknown) => setError(e instanceof ApiError ? e.message : "Enrollment failed"));
  }, []);

  async function onConfirm(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const result = await authService.confirmEnrollment(code);
      setRecoveryCodes(result.recovery_codes);
    } catch (e: unknown) {
      setError(e instanceof ApiError ? e.message : "Confirmation failed");
    } finally {
      setLoading(false);
    }
  }

  if (recoveryCodes) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Save your recovery codes</CardTitle>
          <CardDescription>These are shown once. Store them safely.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <ul className="grid grid-cols-2 gap-2 font-mono text-sm">
            {recoveryCodes.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
          <Button className="w-full" onClick={() => router.push("/")}>
            Continue
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Enroll two-factor authentication</CardTitle>
        <CardDescription>Add this account to your authenticator app</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        {secret ? (
          <>
            <div className="space-y-1">
              <p className="text-sm font-medium">Secret</p>
              <p className="break-all font-mono text-sm">{secret}</p>
            </div>
            {otpauthUri ? (
              <div className="space-y-1">
                <p className="text-sm font-medium">otpauth URI</p>
                <p className="break-all text-xs">{otpauthUri}</p>
              </div>
            ) : null}
            <form className="space-y-4" onSubmit={onConfirm}>
              <div className="space-y-2">
                <Label htmlFor="code">Verification code</Label>
                <Input id="code" value={code} onChange={(e) => setCode(e.target.value)} required />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Confirming..." : "Confirm enrollment"}
              </Button>
            </form>
          </>
        ) : (
          <p className="text-sm">Loading enrollment...</p>
        )}
      </CardContent>
    </Card>
  );
}
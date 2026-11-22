import { MfaVerifyForm } from "@/features/auth/components/mfa-verify-form";

export default function MfaPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-4">
      <MfaVerifyForm />
    </div>
  );
}
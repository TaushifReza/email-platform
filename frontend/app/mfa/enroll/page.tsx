import { MfaEnroll } from "@/features/auth/components/mfa-enroll";

export default function MfaEnrollPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-4">
      <MfaEnroll />
    </div>
  );
}
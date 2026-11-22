import {
  acceptInvitation,
  enrollMFA,
  confirmMFA,
  getSession,
  login,
  logout,
  stepUp,
  verifyMFA,
} from "@/lib/api";

export const authService = {
  login,
  logout,
  session: getSession,
  enroll: enrollMFA,
  confirmEnrollment: confirmMFA,
  verify: verifyMFA,
  stepUp,
  acceptInvitation,
};
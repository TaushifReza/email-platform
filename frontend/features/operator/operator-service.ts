import * as api from "@/lib/api";

export const operatorService = {
  session: () => api.getOperatorSession(),
  listOrganizations: (limit?: number, offset?: number) => api.listOperatorOrganizations(limit, offset),
  createOrganization: (input: {
    name: string;
    defaultSenderName?: string;
    defaultReplyTo?: string;
    timeZone?: string;
  }) => api.createOperatorOrganization(input),
  suspend: (organizationId: string) => api.suspendOrganization(organizationId),
  reinstate: (organizationId: string) => api.reinstateOrganization(organizationId),
  updateLimits: (
    organizationId: string,
    input: {
      storageQuotaBytes: number;
      capsEnabled: boolean;
      dailyEmailCap?: number;
      campaignRecipientCap?: number;
    },
  ) => api.updateOrganizationLimits(organizationId, input),
  listAudit: (limit?: number, offset?: number) => api.listOperatorAudit(limit, offset),
};

export type OperatorOrganization = NonNullable<api.OrganizationList["data"]>[number];
export type OperatorAuditEntry = NonNullable<api.AuditEntryList["data"]>[number];

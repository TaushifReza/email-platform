import {
  getOrganization,
  updateOrganizationPolicy,
  updateOrganizationProfile,
} from "@/lib/api";

export const organizationService = {
  get: getOrganization,
  updateProfile: updateOrganizationProfile,
  updatePolicy: updateOrganizationPolicy,
};
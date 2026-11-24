package audit

// Action identifiers for AUD-01. Stable strings; do not rename.
const (
	ActionSignIn              = "auth.sign_in"
	ActionSignInFailed        = "auth.sign_in_failed"
	ActionSignOut             = "auth.sign_out"
	ActionPasswordResetReq    = "auth.password_reset_requested"
	ActionPasswordResetDone   = "auth.password_reset_completed"
	ActionMFAEnrolled         = "auth.mfa_enrolled"
	ActionMFAVerified         = "auth.mfa_verified"
	ActionMFAFailed           = "auth.mfa_failed"
	ActionOrgCreated          = "org.created"
	ActionOrgSuspended        = "org.suspended"
	ActionOrgReinstated       = "org.reinstated"
	ActionOrgLimitsChanged    = "org.limits_changed"
	ActionOrgProfileUpdated   = "org.profile_updated"
	ActionOrgPolicyUpdated    = "org.policy_updated"
	ActionUserInvited         = "user.invited"
	ActionUserRoleChanged     = "user.role_changed"
	ActionUserDeactivated     = "user.deactivated"
	ActionProviderCreated     = "provider.created"
	ActionProviderUpdated     = "provider.updated"
	ActionProviderRotated     = "provider.rotated"
	ActionProviderDeactivated = "provider.deactivated"
)

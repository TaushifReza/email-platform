package identity

import "errors"

// Role constants. Operator accounts live in a separate table with no organization.
const (
	rolePlatformAdmin   = "platform_admin"
	roleOwner           = "owner"
	roleAdmin           = "admin"
	roleCampaignManager = "campaign_manager"
	roleReviewer        = "reviewer"
	roleViewer          = "viewer"
)

// Capability identifies one row of the SRS 6.2 permission matrix (plus the
// operator pause capability from ADM-01). Unknown values are denied by default.
type Capability string

const (
	CapManageOrganizations      = "org.manage"                 // Platform Admin
	CapOrgLimits                = "org.limits"                 // Platform Admin
	CapOperatorAudit            = "operator.audit"             // Platform Admin
	CapTenantContentView        = "tenant.content.view"        // all tenant roles; PA with grant
	CapProviderCredentialReveal = "provider.credential.reveal" // never
	CapOrgProfileSettings       = "org.profile.settings"       // Owner, Admin
	CapOrgPolicySettings        = "org.policy.settings"        // Owner only (approval requirement, retention)
	CapManageUsers              = "org.users.manage"           // Owner; Admin except Owners
	CapGrantSupportAccess       = "org.support.grant"          // Owner
	CapManageProviders          = "provider.manage"            // Owner, Admin
	CapEditTemplates            = "template.edit"              // Owner, Admin, Campaign Manager
	CapViewTemplates            = "template.view"              // all tenant roles
	CapCreateCampaigns          = "campaign.create"            // Owner, Admin, Campaign Manager
	CapLaunchCampaign           = "campaign.launch"            // Owner, Admin, Campaign Manager
	CapApproveCampaign          = "campaign.approve"           // Owner, Admin, Reviewer; never own
	CapControlCampaign          = "campaign.control"           // Owner, Admin; CM own only; PA denied
	CapAbusePauseCampaign       = "campaign.abuse.pause"       // Platform Admin only (ADM-01)
	CapViewReports              = "report.view"                // all tenant roles; PA with grant
	CapExportReports            = "report.export"              // Owner, Admin, CM; PA with grant
	CapDeleteRecipientData      = "recipient.data.delete"      // Owner, Admin
	CapSuppressionAdd           = "suppression.add"            // Owner, Admin, Campaign Manager
	CapSuppressionRemove        = "suppression.remove"         // Owner, Admin
	CapViewAuditLog             = "audit.view"                 // Owner, Admin
)

// Access is the raw cell value from the matrix.
type Access uint8

const (
	Deny Access = iota
	Allow
	AllowOwn
	AllowWithGrant
)

// Request describes the specific action being authorized. Organization and
// identity always come from the Actor (session), never from Request fields.
type Request struct {
	Capability   Capability
	OwnerUserID  string // creator of the target resource; "" when not applicable
	TargetRole   string // role of the user being managed (CapManageUsers)
	SupportGrant bool   // active Owner-issued grant (M1; false in MVP)
}

// ErrForbidden is returned by Authorize when the actor may not perform the action.
var ErrForbidden = errors.New("identity: forbidden")

// matrix maps (role, capability) -> raw access. Every tenant role and the
// platform admin role must have an entry for every capability; missing entries
// are treated as Deny (fail closed).
var matrix = map[Capability]map[string]Access{ //nolint:dupl,gochecknoglobals
	CapManageOrganizations:      {rolePlatformAdmin: Allow},
	CapOrgLimits:                {rolePlatformAdmin: Allow},
	CapOperatorAudit:            {rolePlatformAdmin: Allow},
	CapTenantContentView:        {rolePlatformAdmin: AllowWithGrant, roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow, roleReviewer: Allow, roleViewer: Allow}, //nolint:lll
	CapProviderCredentialReveal: {},
	CapOrgProfileSettings:       {roleOwner: Allow, roleAdmin: Allow},
	CapOrgPolicySettings:        {roleOwner: Allow},
	CapManageUsers:              {roleOwner: Allow, roleAdmin: Allow},
	CapGrantSupportAccess:       {roleOwner: Allow},
	CapManageProviders:          {roleOwner: Allow, roleAdmin: Allow},
	CapEditTemplates:            {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow},
	CapViewTemplates:            {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow, roleReviewer: Allow, roleViewer: Allow}, //nolint:lll
	CapCreateCampaigns:          {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow},
	CapLaunchCampaign:           {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow},
	CapApproveCampaign:          {roleOwner: Allow, roleAdmin: Allow, roleReviewer: Allow},
	CapControlCampaign:          {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: AllowOwn},
	CapAbusePauseCampaign:       {rolePlatformAdmin: Allow},
	CapViewReports:              {rolePlatformAdmin: AllowWithGrant, roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow, roleReviewer: Allow, roleViewer: Allow}, //nolint:lll
	CapExportReports:            {rolePlatformAdmin: AllowWithGrant, roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow},                                         //nolint:lll
	CapDeleteRecipientData:      {roleOwner: Allow, roleAdmin: Allow},
	CapSuppressionAdd:           {roleOwner: Allow, roleAdmin: Allow, roleCampaignManager: Allow},
	CapSuppressionRemove:        {roleOwner: Allow, roleAdmin: Allow},
	CapViewAuditLog:             {roleOwner: Allow, roleAdmin: Allow},
}

// allCapabilities lists every defined capability.
func AllCapabilities() []Capability {
	return []Capability{
		CapManageOrganizations, CapOrgLimits, CapOperatorAudit, CapTenantContentView,
		CapProviderCredentialReveal, CapOrgProfileSettings, CapOrgPolicySettings,
		CapManageUsers, CapGrantSupportAccess, CapManageProviders, CapEditTemplates,
		CapViewTemplates, CapCreateCampaigns, CapLaunchCampaign, CapApproveCampaign,
		CapControlCampaign, CapAbusePauseCampaign, CapViewReports, CapExportReports,
		CapDeleteRecipientData, CapSuppressionAdd, CapSuppressionRemove, CapViewAuditLog,
	}
}

// Authorize applies the matrix plus scope rules. It returns nil on allow.
func Authorize(actor Actor, req Request) error {
	access, ok := matrix[req.Capability][actor.Role]
	if !ok {
		return ErrForbidden
	}
	switch access {
	case Deny:
		return ErrForbidden
	case Allow:
		// no further scope check
	case AllowWithGrant:
		if !req.SupportGrant {
			return ErrForbidden
		}
	case AllowOwn:
		if req.OwnerUserID == "" || req.OwnerUserID != actor.UserID {
			return ErrForbidden
		}
	}
	// Scope rules that refine the matrix cells.
	switch req.Capability {
	case CapApproveCampaign:
		// Approver must not be the creator of the campaign.
		if req.OwnerUserID != "" && req.OwnerUserID == actor.UserID {
			return ErrForbidden
		}
	case CapManageUsers:
		// Admins may not manage Owners.
		if actor.Role == roleAdmin && req.TargetRole == roleOwner {
			return ErrForbidden
		}
	case CapControlCampaign:
		// Platform Admin is handled via CapAbusePauseCampaign; never here.
		// Campaign Managers are already restricted to AllowOwn above.
	case CapAbusePauseCampaign:
		// Only Platform Admin; already encoded in the matrix.
	}
	return nil
}

// validRoles is the set of roles a user account may hold.
func validRole(role string) bool {
	switch role {
	case roleOwner, roleAdmin, roleCampaignManager, roleReviewer, roleViewer:
		return true
	}
	return false
}

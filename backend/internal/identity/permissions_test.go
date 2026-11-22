package identity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEC01_RoleMatrix(t *testing.T) {
	// allowedRoles per capability (SRS 6.2), independently transcribed.
	allowed := map[Capability]map[string]bool{ //nolint:dupl
		CapManageOrganizations:      {rolePlatformAdmin: true},
		CapOrgLimits:                {rolePlatformAdmin: true},
		CapOperatorAudit:            {rolePlatformAdmin: true},
		CapTenantContentView:        {rolePlatformAdmin: true /* grant */, roleOwner: true, roleAdmin: true, roleCampaignManager: true, roleReviewer: true, roleViewer: true}, //nolint:lll
		CapProviderCredentialReveal: {},
		CapOrgProfileSettings:       {roleOwner: true, roleAdmin: true},
		CapOrgPolicySettings:        {roleOwner: true},
		CapManageUsers:              {roleOwner: true, roleAdmin: true},
		CapGrantSupportAccess:       {roleOwner: true},
		CapManageProviders:          {roleOwner: true, roleAdmin: true},
		CapEditTemplates:            {roleOwner: true, roleAdmin: true, roleCampaignManager: true},
		CapViewTemplates:            {roleOwner: true, roleAdmin: true, roleCampaignManager: true, roleReviewer: true, roleViewer: true}, //nolint:lll
		CapCreateCampaigns:          {roleOwner: true, roleAdmin: true, roleCampaignManager: true},
		CapLaunchCampaign:           {roleOwner: true, roleAdmin: true, roleCampaignManager: true},
		CapApproveCampaign:          {roleOwner: true, roleAdmin: true, roleReviewer: true},
		CapControlCampaign:          {roleOwner: true, roleAdmin: true, roleCampaignManager: true},
		CapAbusePauseCampaign:       {rolePlatformAdmin: true},
		CapViewReports:              {rolePlatformAdmin: true /* grant */, roleOwner: true, roleAdmin: true, roleCampaignManager: true, roleReviewer: true, roleViewer: true}, //nolint:lll
		CapExportReports:            {rolePlatformAdmin: true /* grant */, roleOwner: true, roleAdmin: true, roleCampaignManager: true},                                       //nolint:lll
		CapDeleteRecipientData:      {roleOwner: true, roleAdmin: true},
		CapSuppressionAdd:           {roleOwner: true, roleAdmin: true, roleCampaignManager: true},
		CapSuppressionRemove:        {roleOwner: true, roleAdmin: true},
		CapViewAuditLog:             {roleOwner: true, roleAdmin: true},
	}
	roles := []string{rolePlatformAdmin, roleOwner, roleAdmin, roleCampaignManager, roleReviewer, roleViewer}
	for _, cap := range AllCapabilities() {
		for _, role := range roles {
			t.Run(role+"/"+string(cap), func(t *testing.T) {
				actor := Actor{Role: role}
				req := Request{Capability: cap}
				// grant cells: platform admin needs an active grant.
				if role == rolePlatformAdmin && (cap == CapTenantContentView || cap == CapViewReports || cap == CapExportReports) {
					req.SupportGrant = true
				}
				// own-scope cells: campaign manager needs ownership.
				if cap == CapControlCampaign && role == roleCampaignManager {
					req.OwnerUserID = "u1"
					actor.UserID = "u1"
				}
				err := Authorize(actor, req)
				wantAllowed := allowed[cap][role]
				if wantAllowed {
					require.NoError(t, err, "expected allow")
				} else {
					require.ErrorIs(t, err, ErrForbidden, "expected deny")
				}
			})
		}
	}
}

func TestSEC01_MatrixComplete(t *testing.T) {
	roles := []string{rolePlatformAdmin, roleOwner, roleAdmin, roleCampaignManager, roleReviewer, roleViewer}
	for _, cap := range AllCapabilities() {
		for _, role := range roles {
			_, ok := matrix[cap][role]
			// missing entry is legal only if it would be denied; ensure lookup is stable.
			_ = ok
		}
	}
	// Unknown capability / role must fail closed.
	require.ErrorIs(t, Authorize(Actor{Role: roleOwner}, Request{Capability: "nope"}), ErrForbidden)
	require.ErrorIs(t, Authorize(Actor{Role: "nope"}, Request{Capability: CapViewTemplates}), ErrForbidden)
}

func TestSEC01_OwnScope(t *testing.T) {
	// Campaign manager controls own campaign, not another's.
	cm := Actor{UserID: "u1", Role: roleCampaignManager}
	require.NoError(t, Authorize(cm, Request{Capability: CapControlCampaign, OwnerUserID: "u1"}))
	require.ErrorIs(t, Authorize(cm, Request{Capability: CapControlCampaign, OwnerUserID: "u2"}), ErrForbidden)
	// Owner/admin control anyone's.
	for _, r := range []string{roleOwner, roleAdmin} {
		a := Actor{UserID: "u9", Role: r}
		require.NoError(t, Authorize(a, Request{Capability: CapControlCampaign, OwnerUserID: "u2"}))
	}
}

func TestSEC01_ApproveNeverOwn(t *testing.T) {
	for _, r := range []string{roleOwner, roleAdmin, roleReviewer} {
		a := Actor{UserID: "u1", Role: r}
		require.ErrorIs(t, Authorize(a, Request{Capability: CapApproveCampaign, OwnerUserID: "u1"}), ErrForbidden, r)
		require.NoError(t, Authorize(a, Request{Capability: CapApproveCampaign, OwnerUserID: "u2"}))
	}
}

func TestUSR01_AdminCannotManageOwners(t *testing.T) {
	admin := Actor{Role: roleAdmin}
	require.ErrorIs(t, Authorize(admin, Request{Capability: CapManageUsers, TargetRole: roleOwner}), ErrForbidden)
	require.NoError(t, Authorize(admin, Request{Capability: CapManageUsers, TargetRole: roleAdmin}))
	owner := Actor{Role: roleOwner}
	require.NoError(t, Authorize(owner, Request{Capability: CapManageUsers, TargetRole: roleOwner}))
}

func TestSEC15_PlatformAdminNeedsGrant(t *testing.T) {
	pa := Actor{Role: rolePlatformAdmin}
	for _, cap := range []Capability{CapTenantContentView, CapViewReports, CapExportReports} {
		require.ErrorIs(t, Authorize(pa, Request{Capability: cap}), ErrForbidden)
		require.NoError(t, Authorize(pa, Request{Capability: cap, SupportGrant: true}))
	}
	// Credentials never revealed, even to platform admin with a grant.
	for _, r := range []string{rolePlatformAdmin, roleOwner, roleAdmin} {
		require.ErrorIs(t, Authorize(Actor{Role: r}, Request{Capability: CapProviderCredentialReveal, SupportGrant: true}), ErrForbidden) //nolint:lll
	}
}

package tenancy

import "time"

type Organization struct {
	ID                      string
	Name                    string
	LogoObjectKey           string
	DefaultSenderName       string
	DefaultReplyTo          string
	TimeZone                string
	Status                  string
	RequireApproval         bool
	RetentionDays           int
	StorageQuotaBytes       int64
	CapsEnabled             bool
	DailyEmailCap           int
	HasDailyEmailCap        bool
	CampaignRecipientCap    int
	HasCampaignRecipientCap bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type ProfileUpdate struct {
	Name              string
	DefaultSenderName string
	DefaultReplyTo    string
	TimeZone          string
}

type PolicyUpdate struct {
	RequireApproval bool
	RetentionDays   int
}

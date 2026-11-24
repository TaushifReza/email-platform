package smtpclient

import "testing"

func TestSND_ReplyClassificationTable(t *testing.T) {
	cases := []struct {
		stage Stage
		code  int
		reply string
		want  Class
	}{
		{StageBody, 250, "250 OK", Accepted},
		{StageRcpt, 421, "421 try later", Temporary},
		{StageRcpt, 450, "450 mailbox unavailable", Temporary},
		{StageRcpt, 451, "451 local error", Temporary},
		{StageRcpt, 452, "452 insufficient storage", Temporary},
		{StageRcpt, 400, "4.2.2 mailbox full", Temporary},
		{StageConnect, 0, "connection refused", TemporaryConnection},
		{StageRcpt, 550, "550 no such user", PermanentRecipient},
		{StageRcpt, 551, "551 not local", PermanentRecipient},
		{StageRcpt, 553, "553 mailbox name invalid", PermanentRecipient},
		{StageData, 552, "552 message too large", PermanentMessage},
		{StageData, 554, "554 transaction failed", PermanentMessage},
		{StageAuth, 530, "530 auth required", ProviderAuth},
		{StageAuth, 535, "535 bad credentials", ProviderAuth},
		{StageMailFrom, 550, "550 sender rejected", ProviderPolicy},
		{StageData, 0, "5.7.1 blocked", ProviderPolicy},
		{StageData, 0, "sender blocked by reputation", ProviderPolicy},
		{StageData, 0, "provider daily quota exceeded", ProviderQuota},
		{StageRcpt, 999, "weird", Temporary},
	}
	for _, c := range cases {
		got := classify(c.stage, c.code, c.reply)
		if got != c.want {
			t.Errorf("stage=%d code=%d reply=%q: got %d want %d", c.stage, c.code, c.reply, got, c.want)
		}
	}
}

func TestSND_UnknownNeverSuccess(t *testing.T) {
	for _, st := range []Stage{StageConnect, StageAuth, StageMailFrom, StageRcpt, StageData, StageBody} {
		if got := classify(st, 0, "totally unknown"); got == Accepted {
			t.Fatalf("stage %d: unknown became Accepted", st)
		}
	}
}

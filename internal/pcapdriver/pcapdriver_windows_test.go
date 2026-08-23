//go:build windows

package pcapdriver

import "testing"

func TestDecide_NpcapAllUsers_NoWarning(t *testing.T) {
	got := decide(true, false, true, false)
	if got != (Warning{}) {
		t.Fatalf("decide(npcap present, all-users) = %+v, want zero Warning", got)
	}
}

func TestDecide_NpcapAdminOnly_ProcessElevated_NoWarning(t *testing.T) {
	got := decide(true, true, false, true)
	if got != (Warning{}) {
		t.Fatalf("decide(npcap admin-only, elevated) = %+v, want zero Warning", got)
	}
}

func TestDecide_NpcapAdminOnly_ProcessNotElevated_Warns(t *testing.T) {
	got := decide(true, true, false, false)
	if got.Message == "" {
		t.Fatal("decide(npcap admin-only, not elevated) returned no warning, want one telling the user to restart as Administrator")
	}
	if got.HelpURL != "" {
		t.Fatalf("decide(npcap admin-only, not elevated) HelpURL = %q, want empty - the fix is restarting elevated, not a download", got.HelpURL)
	}
}

func TestDecide_OnlyWinPcap_WarnsWithDownloadLink(t *testing.T) {
	got := decide(false, false, true, false)
	if got.Message == "" {
		t.Fatal("decide(only WinPcap present) returned no warning, want one")
	}
	if got.HelpURL != npcapDownloadURL {
		t.Fatalf("decide(only WinPcap present) HelpURL = %q, want %q", got.HelpURL, npcapDownloadURL)
	}
}

func TestDecide_NeitherDriverPresent_WarnsWithDownloadLink(t *testing.T) {
	got := decide(false, false, false, false)
	if got.Message == "" {
		t.Fatal("decide(no driver present) returned no warning, want one")
	}
	if got.HelpURL != npcapDownloadURL {
		t.Fatalf("decide(no driver present) HelpURL = %q, want %q", got.HelpURL, npcapDownloadURL)
	}
}

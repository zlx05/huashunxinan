package engine

import "strings"

// osTokens is the cross-protocol OS hint dictionary, used as a backstop when
// a rule's specific OS regex does not produce a hint.
var osTokens = []string{
	"Ubuntu", "Debian", "CentOS", "Red Hat", "RedHat", "Fedora",
	"FreeBSD", "OpenBSD", "Alpine", "Amazon", "SUSE", "Arch",
}

// detectOS scans a banner for a known OS token and returns it (empty if none).
func detectOS(banner string) string {
	for _, tok := range osTokens {
		if strings.Contains(banner, tok) {
			return tok
		}
	}
	return ""
}

// Package passwords implements personal & commercial password entries with
// encryption at rest, favicon resolution and reveal auditing.
package passwords

import (
	"net/url"
	"strings"
)

// ExtractDomain returns the bare hostname (without a leading "www.") from a
// site URL. It returns "" when the input is empty or not http/https.
func ExtractDomain(siteURL string) string {
	siteURL = strings.TrimSpace(siteURL)
	if siteURL == "" {
		return ""
	}
	// Allow bare hostnames like "github.com" by adding a scheme for parsing.
	if !strings.Contains(siteURL, "://") {
		siteURL = "https://" + siteURL
	}
	u, err := url.Parse(siteURL)
	if err != nil {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	host := u.Hostname()
	host = strings.ToLower(host)
	host = strings.TrimPrefix(host, "www.")
	return host
}

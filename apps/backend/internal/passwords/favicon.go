package passwords

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// faviconClient has a short timeout so a slow site never blocks a request.
var faviconClient = &http.Client{Timeout: 3 * time.Second}

// googleFaviconURL returns Google's favicon service URL for a domain. It almost
// always renders something (falling back to a generic globe), which makes it a
// dependable last resort when a site does not expose /favicon.ico directly.
func googleFaviconURL(domain string) string {
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", domain)
}

// ResolveFavicon returns the best favicon URL for a domain. It first probes
// https://{domain}/favicon.ico and uses it when it responds 200 with an image
// content-type; otherwise it falls back to Google's favicon service so an icon
// is always available. ok is false only when the domain is empty.
func ResolveFavicon(ctx context.Context, domain string) (string, bool) {
	if domain == "" {
		return "", false
	}

	if direct := "https://" + domain + "/favicon.ico"; faviconIsImage(ctx, direct) {
		return direct, true
	}
	return googleFaviconURL(domain), true
}

// faviconIsImage reports whether the URL responds 200 with an image content-type.
func faviconIsImage(ctx context.Context, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := faviconClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return strings.HasPrefix(resp.Header.Get("Content-Type"), "image/")
}

// maxCustomIconBytes caps the size of an uploaded custom icon (data URL length).
const maxCustomIconBytes = 256 * 1024

// validCustomIcon reports whether s is an acceptable image data URL small enough
// to store inline. Uploaded icons are sent by the client as base64 data URLs.
func validCustomIcon(s string) bool {
	if s == "" || len(s) > maxCustomIconBytes {
		return false
	}
	return strings.HasPrefix(s, "data:image/")
}

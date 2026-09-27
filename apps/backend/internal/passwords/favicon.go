package passwords

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// faviconClient has a short timeout so a slow site never blocks a request.
var faviconClient = &http.Client{Timeout: 3 * time.Second}

// ResolveFavicon tries https://{domain}/favicon.ico. If it responds 200 with an
// image content-type, the URL is returned. Otherwise ok is false and the caller
// should fall back to the group icon.
func ResolveFavicon(ctx context.Context, domain string) (string, bool) {
	if domain == "" {
		return "", false
	}
	faviconURL := "https://" + domain + "/favicon.ico"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, faviconURL, nil)
	if err != nil {
		return "", false
	}
	resp, err := faviconClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return "", false
	}
	return faviconURL, true
}

package passwords

import (
	"context"
	"sort"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/db"
)

// UnusedThreshold is how long a record may go unused before the health check
// flags it as stale.
const UnusedThreshold = 90 * 24 * time.Hour

// ExpiringWindow is how far ahead the health check looks for secrets that are
// about to expire.
const ExpiringWindow = 30 * 24 * time.Hour

// InventoryItem is a per-type count for the inventory screen.
type InventoryItem struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// Inventory summarises the whole secret estate for the inventory screen.
type Inventory struct {
	Total             int             `json:"total"`
	ByType            []InventoryItem `json:"by_type"`
	ExpiringThisMonth int             `json:"expiring_this_month"`
	ExpiringSoon      int             `json:"expiring_soon"`
	Expired           int             `json:"expired"`
	WeakPasswords     int             `json:"weak_passwords"`
	WithoutOwner      int             `json:"without_owner"`
	Unused            int             `json:"unused"`
}

// Health groups the records that need attention, plus a summary of counts.
type Health struct {
	Summary  HealthSummary `json:"summary"`
	Weak     []View        `json:"weak"`
	NoOwner  []View        `json:"no_owner"`
	Unused   []View        `json:"unused"`
	Expiring []View        `json:"expiring"`
	Expired  []View        `json:"expired"`
}

// HealthSummary holds the issue counts shown as headline numbers.
type HealthSummary struct {
	Total    int `json:"total"`
	Weak     int `json:"weak"`
	NoOwner  int `json:"no_owner"`
	Unused   int `json:"unused"`
	Expiring int `json:"expiring"`
	Expired  int `json:"expired"`
}

// Inventory returns a classification + expiration summary for the owner.
func (s *Service) Inventory(ctx context.Context, ownerID string) (*Inventory, error) {
	entries, err := s.store.ListPasswordEntries(ctx, ownerID, db.PasswordFilter{})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	endOfMonth := endOfMonth(now)
	soonCutoff := now.Add(ExpiringWindow)

	inv := &Inventory{Total: len(entries)}
	counts := map[string]int{}

	for i := range entries {
		e := &entries[i]
		typ := e.EntryType
		if typ == "" {
			typ = "password"
		}
		counts[typ]++

		if e.ExpiresAt != nil {
			switch {
			case e.ExpiresAt.Before(now):
				inv.Expired++
			default:
				if !e.ExpiresAt.After(soonCutoff) {
					inv.ExpiringSoon++
				}
				if !e.ExpiresAt.After(endOfMonth) {
					inv.ExpiringThisMonth++
				}
			}
		}
		if isWeak(e) {
			inv.WeakPasswords++
		}
		if !hasOwner(e) {
			inv.WithoutOwner++
		}
		if isUnused(e, now) {
			inv.Unused++
		}
	}

	inv.ByType = make([]InventoryItem, 0, len(counts))
	for typ, c := range counts {
		inv.ByType = append(inv.ByType, InventoryItem{Type: typ, Count: c})
	}
	// Stable, deterministic ordering: highest count first, then type name.
	sort.Slice(inv.ByType, func(a, b int) bool {
		if inv.ByType[a].Count != inv.ByType[b].Count {
			return inv.ByType[a].Count > inv.ByType[b].Count
		}
		return inv.ByType[a].Type < inv.ByType[b].Type
	})

	return inv, nil
}

// Health returns the records that need attention for the health screen.
func (s *Service) Health(ctx context.Context, ownerID string) (*Health, error) {
	entries, err := s.store.ListPasswordEntries(ctx, ownerID, db.PasswordFilter{})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	soonCutoff := now.Add(ExpiringWindow)

	h := &Health{
		Weak:     []View{},
		NoOwner:  []View{},
		Unused:   []View{},
		Expiring: []View{},
		Expired:  []View{},
	}
	h.Summary.Total = len(entries)

	for i := range entries {
		e := &entries[i]
		v := toView(e)
		if isWeak(e) {
			h.Weak = append(h.Weak, v)
		}
		if !hasOwner(e) {
			h.NoOwner = append(h.NoOwner, v)
		}
		if isUnused(e, now) {
			h.Unused = append(h.Unused, v)
		}
		if e.ExpiresAt != nil {
			if e.ExpiresAt.Before(now) {
				h.Expired = append(h.Expired, v)
			} else if !e.ExpiresAt.After(soonCutoff) {
				h.Expiring = append(h.Expiring, v)
			}
		}
	}

	h.Summary.Weak = len(h.Weak)
	h.Summary.NoOwner = len(h.NoOwner)
	h.Summary.Unused = len(h.Unused)
	h.Summary.Expiring = len(h.Expiring)
	h.Summary.Expired = len(h.Expired)

	return h, nil
}

// isWeak reports whether an entry has a known, weak password strength.
func isWeak(e *db.PasswordEntry) bool {
	return e.PasswordStrength != nil && *e.PasswordStrength <= WeakStrengthThreshold
}

// hasOwner reports whether an entry has a non-empty owner.
func hasOwner(e *db.PasswordEntry) bool {
	return e.Owner != nil && *e.Owner != ""
}

// isUnused reports whether an entry has not been used for longer than the
// threshold (counting from last use, or from creation when never used).
func isUnused(e *db.PasswordEntry, now time.Time) bool {
	ref := e.CreatedAt
	if e.LastUsedAt != nil {
		ref = *e.LastUsedAt
	}
	return now.Sub(ref) > UnusedThreshold
}

// endOfMonth returns the last instant of the calendar month containing t.
func endOfMonth(t time.Time) time.Time {
	firstOfNext := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
	return firstOfNext.Add(-time.Nanosecond)
}

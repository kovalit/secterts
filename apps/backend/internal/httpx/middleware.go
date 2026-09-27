package httpx

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

// RequestID assigns each request a unique id and echoes it in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom returns the request id stored in the context, if any.
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// Recoverer converts panics into 500 responses instead of crashing the server.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered [%s]: %v", RequestIDFrom(r.Context()), rec)
				Error(w, ErrInternal("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger logs method, path, status and duration for each request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

// ClientIP extracts the client IP, honouring X-Forwarded-For behind a proxy.
func ClientIP(r *http.Request) net.IP {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := net.ParseIP(strings.TrimSpace(parts[0])); ip != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return net.ParseIP(host)
}

// RateLimiter is a very small fixed-window in-memory limiter keyed by a string.
// It is deliberately simple — enough for auth endpoints on a single-node MVP.
type RateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	limit    int
	window   time.Duration
	lastGC   time.Time
	gcPeriod time.Duration
}

// NewRateLimiter allows `limit` requests per `window` per key.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:     make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		gcPeriod: window * 4,
		lastGC:   time.Now(),
	}
}

// Allow records a hit for key and reports whether it is within the limit.
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-rl.window)
	recent := rl.hits[key][:0]
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	rl.hits[key] = recent

	if now.Sub(rl.lastGC) > rl.gcPeriod {
		rl.gc(cutoff)
		rl.lastGC = now
	}

	if len(recent) >= rl.limit {
		return false
	}
	rl.hits[key] = append(rl.hits[key], now)
	return true
}

func (rl *RateLimiter) gc(cutoff time.Time) {
	for key, times := range rl.hits {
		keep := times[:0]
		for _, t := range times {
			if t.After(cutoff) {
				keep = append(keep, t)
			}
		}
		if len(keep) == 0 {
			delete(rl.hits, key)
		} else {
			rl.hits[key] = keep
		}
	}
}

// Middleware limits requests keyed by client IP, responding 429 when exceeded.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path + "|"
		if ip := ClientIP(r); ip != nil {
			key += ip.String()
		}
		if !rl.Allow(key) {
			Error(w, ErrTooMany("too many requests, slow down"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

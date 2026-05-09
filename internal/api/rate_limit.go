package api

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

type rateLimiter struct {
	mu        sync.Mutex
	clients   map[string]*rateBucket
	limit     int
	window    time.Duration
	lastPrune time.Time
}

type rateBucket struct {
	count      int
	windowFrom time.Time
	lastSeen   time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		clients:   make(map[string]*rateBucket),
		limit:     limit,
		window:    window,
		lastPrune: time.Now(),
	}
}

func (l *rateLimiter) allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastPrune) > l.window {
		for clientKey, bucket := range l.clients {
			if now.Sub(bucket.lastSeen) > 2*l.window {
				delete(l.clients, clientKey)
			}
		}
		l.lastPrune = now
	}

	bucket := l.clients[key]
	if bucket == nil || now.Sub(bucket.windowFrom) > l.window {
		l.clients[key] = &rateBucket{count: 1, windowFrom: now, lastSeen: now}
		return true
	}

	bucket.lastSeen = now
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	return true
}

func (s *Server) authRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// убрать потом --------------------------------------------------------------------------
		log.Printf("🔍 INCOMING: %s %s | RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		// убрать потом --------------------------------------------------------------------------
		key := remoteIP(r) + ":" + r.URL.Path
		if s.authLimiter != nil && !s.authLimiter.allow(key) {
			log.Printf("auth rate limit exceeded: request_id=%s ip=%s path=%s", requestIDFromContext(r.Context()), remoteIP(r), r.URL.Path)
			writeError(w, r, app.RateLimited())
			return
		}
		next.ServeHTTP(w, r)
	})
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

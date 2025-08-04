package auth

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type attempt struct {
	count int
	until time.Time
}

// LimitLogin bounds password attempts without trusting forwarded IP headers.
func LimitLogin(next http.HandlerFunc) http.HandlerFunc {
	var mu sync.Mutex
	attempts := map[string]attempt{}
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		now := time.Now()
		mu.Lock()
		if len(attempts) > 10000 {
			for key, value := range attempts {
				if now.After(value.until) {
					delete(attempts, key)
				}
			}
		}
		current := attempts[host]
		if now.After(current.until) {
			current = attempt{until: now.Add(time.Minute)}
		}
		if current.count >= 10 || (len(attempts) > 10000 && current.count == 0) {
			mu.Unlock()
			w.Header().Set("Retry-After", "60")
			http.Error(w, "too many login attempts", 429)
			return
		}
		current.count++
		attempts[host] = current
		mu.Unlock()
		next(w, r)
	}
}

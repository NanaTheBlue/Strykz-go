package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors sync.Map
	rate     int
	duration time.Duration
}

type Visitor struct {
	mu        sync.Mutex
	tokens    int
	lastCheck time.Time
}

func NewRateLimiter(rate int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:     rate,
		duration: duration,
	}
	go rl.cleanup()
	return rl
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		return r.RemoteAddr
	}

	forwardedIPs := strings.Split(ip, ",")
	return strings.TrimSpace(forwardedIPs[0])
}

func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		v, _ := rl.visitors.LoadOrStore(ip, &Visitor{
			tokens:    rl.rate,
			lastCheck: time.Now(),
		})

		visitor := v.(*Visitor)
		visitor.mu.Lock()

		now := time.Now()
		elapsed := now.Sub(visitor.lastCheck)

		refill := int(elapsed.Seconds() / rl.duration.Seconds() * float64(rl.rate))
		if refill > 0 {
			visitor.tokens += refill
			if visitor.tokens > rl.rate {
				visitor.tokens = rl.rate
			}
			visitor.lastCheck = now
		}

		if visitor.tokens <= 0 {
			visitor.mu.Unlock()
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}

		visitor.tokens--
		visitor.mu.Unlock()

		next(w, r)
	}
}
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)

		now := time.Now()
		rl.visitors.Range(func(key, value any) bool {
			v := value.(*Visitor)
			v.mu.Lock()
			if now.Sub(v.lastCheck) > 10*time.Minute {
				rl.visitors.Delete(key)
			}
			v.mu.Unlock()
			return true
		})
	}
}

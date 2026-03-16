package middleware

import (
	"net/http"
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
	return rl
}

func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

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
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}

		visitor.tokens--
		visitor.mu.Unlock()

		next(w, r)
	}
}

package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
    mu       sync.Mutex
    limiters = map[string]*rate.Limiter{}
)

func getLimiter(ip string) *rate.Limiter {
    mu.Lock(); defer mu.Unlock()
    if l, ok := limiters[ip]; ok { return l }
    l := rate.NewLimiter(rate.Every(time.Second), 5) // 5 req/s por IP
    limiters[ip] = l
    return l
}

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        if !getLimiter(ip).Allow() {
            http.Error(w, "too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
package limiter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Limiter struct {
	storage   Storage
	rateIP    int
	rateToken int
	blockTime time.Duration
}

func NewLimiter(storage Storage, rateIP, rateToken int, blockTime time.Duration) *Limiter {
	return &Limiter{
		storage:   storage,
		rateIP:    rateIP,
		rateToken: rateToken,
		blockTime: blockTime,
	}
}

func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		token := c.GetHeader("API_KEY")

		allowed, err := l.AllowRequest(ip, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "rate limiter error",
				"message": "internal server error",
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			return
		}

		c.Next()
	}
}

func (l *Limiter) AllowRequest(ip, token string) (bool, error) {
	// Verificar bloqueio por IP
	if blocked, err := l.storage.IsBlocked(ip); err != nil || blocked {
		return false, err
	}

	// Verificar bloqueio por Token
	if token != "" {
		if blocked, err := l.storage.IsBlocked(token); err != nil || blocked {
			return false, err
		}
	}

	// Aplicar limite por IP
	count, err := l.storage.Increment(ip)
	if err != nil {
		return false, err
	}
	if count > l.rateIP {
		_ = l.storage.Block(ip, l.blockTime)
		return false, nil
	}

	// Aplicar limite por Token
	if token != "" {
		count, err = l.storage.Increment(token)
		if err != nil {
			return false, err
		}
		if count > l.rateToken {
			_ = l.storage.Block(token, l.blockTime)
			return false, nil
		}
	}

	return true, nil
}

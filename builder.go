package ginratelimiter

import (
	"fmt"
	"ginratelimiter/ratelimit"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Builder 用于构建限流器中间件
type Builder struct {
	prefix  string            // 前缀，用于唯一标识
	limiter ratelimit.Limiter // 限流器接口
}

// NewBuilder 构造函数，用于创建 Builder 实例
func NewBuilder(prefix string, limiter ratelimit.Limiter) *Builder {
	return &Builder{
		prefix:  prefix,
		limiter: limiter,
	}
}

// Build 创建 Gin 中间件，用于限流
func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			// 若发生错误，返回内部服务错误
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			// 超出限流阈值，返回 429 Too Many Requests
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next() // 继续处理请求
	}
}

// limit 方法构建限流键并调用限流接口
// 默认使用 IP限流
func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limiter.Limit(ctx, key)
}

package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisSlidingWindowLimiter 使用 Redis 滑动窗口算法实现限流
type RedisSlidingWindowLimiter struct {
	cmd      redis.Cmdable // Redis 客户端接口
	interval time.Duration // 窗口大小
	rate     int           // 阈值
}

//go:embed slide_window.lua
var luaSlideWindow string // 嵌入 Lua 脚本

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

// Limit 方法检查请求是否超过限流阈值
func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}

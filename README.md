## Gin Sliding Window Rate Limiter
基于 Redis 和 Lua 的 Gin 框架滑动窗口限流中间件，支持高效限流，适用于高并发场景。

### 功能简介
- 实现滑动窗口算法的限流中间件，适用于 Gin 框架。 
- 使用 Redis 作为数据存储，通过 Lua 脚本实现滑动窗口算法。 
- 提供了灵活的限流配置，可自定义限流窗口大小和请求限制数。

### 特性
- 高效：通过 Redis Lua 脚本保证原子性操作，减少网络延迟。
- 灵活：支持自定义限流阈值和时间窗口，满足不同应用场景。
- 易用：集成简单，可作为中间件直接应用于 Gin 项目。

### 使用示例
1. 引入中间件
~~~go
import (
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v9"
    "github.com/nanxu0124/ginratelimiter"
)

func main() {
    // 初始化 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // 创建限流器和 Builder
    limiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, time.Second, 10) // 1秒内允许10次请求
    builder := NewBuilder("api", limiter)

    // 初始化 Gin 实例并应用中间件
    router := gin.Default()
    router.Use(builder.Build())

    // 路由示例
    router.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })

    router.Run(":8080")
}
~~~

2. 配置说明
- interval：滑动窗口大小（时间段）。
- rate：在 interval 时间段内允许的最大请求数。

### 项目结构
~~~text
ginratelimiter/
├── builder.go               # 中间件构建器，用于构建限流中间件
├── ratelimit/
│   ├── types.go             # 限流接口定义
│   ├── redis_slide_window.go # 基于 Redis 和 Lua 实现的滑动窗口限流器
│   └── slide_window.lua     # Lua 脚本，限流逻辑
└── README.md                # 项目说明文档
~~~

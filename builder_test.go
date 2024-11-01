package ginratelimiter

import (
	"context"
	"ginratelimiter/ratelimit"
	"ginratelimiter/ratelimit/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuilder_Build_NotLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建一个 mock 的 Limiter 实例，模拟不触发限流
	mockLimiter := mocks.NewMockLimiter(ctrl)
	mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

	// 创建 Builder 实例
	builder := NewBuilder("test_prefix", mockLimiter)
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Allowed", w.Body.String())
}

func TestBuilder_Build_Limited(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建一个 mock 的 Limiter 实例，模拟触发限流
	mockLimiter := mocks.NewMockLimiter(ctrl)
	mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)

	// 创建 Builder 实例
	builder := NewBuilder("test_prefix", mockLimiter)
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEqual(t, "Allowed", w.Body.String())
}

func TestBuilder_Build_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建一个 mock 的 Limiter 实例，模拟限流器返回错误
	mockLimiter := mocks.NewMockLimiter(ctrl)
	mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, assert.AnError).Times(1)

	// 创建 Builder 实例
	builder := NewBuilder("test_prefix", mockLimiter)
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// 初始化 Redis 客户端
func setupRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis 地址
	})
}

func TestBuilder_Middleware_NotLimited(t *testing.T) {
	rdb := setupRedisClient()
	defer rdb.FlushDB(context.Background()) // 清空 Redis 数据库，避免影响其他测试

	// 创建限流器和 Builder 实例
	limiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, time.Second, 5) // 1 秒内允许 5 次请求
	builder := NewBuilder("test_prefix", limiter)

	// 创建 Gin 实例并应用限流中间件
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 在限流阈值内的请求应返回 200 OK
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Allowed", w.Body.String())
	}
}

func TestBuilder_Middleware_Limited(t *testing.T) {
	rdb := setupRedisClient()
	defer rdb.FlushDB(context.Background())

	// 创建限流器和 Builder 实例
	limiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, time.Second, 3) // 1 秒内允许 3 次请求
	builder := NewBuilder("test_prefix", limiter)

	// 创建 Gin 实例并应用限流中间件
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 前 3 次请求应返回 200 OK
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Allowed", w.Body.String())
	}

	// 第 4 次请求应返回 429 Too Many Requests
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestBuilder_Middleware_ResetAfterInterval(t *testing.T) {
	rdb := setupRedisClient()
	defer rdb.FlushDB(context.Background())

	// 创建限流器和 Builder 实例
	limiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, 500*time.Millisecond, 2) // 500 毫秒窗口，允许 2 次请求
	builder := NewBuilder("test_prefix", limiter)

	// 创建 Gin 实例并应用限流中间件
	router := gin.Default()
	router.Use(builder.Build())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})

	// 发送两次请求，不应被限制
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Allowed", w.Body.String())
	}

	// 等待窗口过期
	time.Sleep(600 * time.Millisecond)

	// 第三次请求应当不被限制
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Allowed", w.Body.String())
}

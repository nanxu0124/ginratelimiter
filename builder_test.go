package ginratelimiter

import (
	"ginratelimiter/ratelimit/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

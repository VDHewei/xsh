package executor

import (
	"fmt"
	"math"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int
	RetryDelay    time.Duration
	BackoffFactor float64
	MaxDelay      time.Duration
}

// DefaultRetryConfig 默认重试配置: 3次, 1s间隔, 2x退避, 最大10s
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
	}
}

// NewRetryConfig 创建自定义重试配置
func NewRetryConfig(maxRetries int, delay time.Duration, backoff float64) *RetryConfig {
	return &RetryConfig{
		MaxRetries:    maxRetries,
		RetryDelay:    delay,
		BackoffFactor: backoff,
		MaxDelay:      30 * time.Second,
	}
}

// Do 执行带重试的函数
func (c *RetryConfig) Do(fn func() (string, error)) (string, error) {
	var lastErr error
	currentDelay := c.RetryDelay

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			// 等待后再重试
			time.Sleep(currentDelay)
			currentDelay = time.Duration(math.Min(
				float64(currentDelay)*c.BackoffFactor,
				float64(c.MaxDelay),
			))
		}

		result, err := fn()
		if err == nil {
			if attempt > 0 {
				result = fmt.Sprintf("%s [retry:%d]", result, attempt)
			}
			return result, nil
		}

		lastErr = err

		// 如果是最后一次尝试, 返回错误
		if attempt == c.MaxRetries {
			return "", fmt.Errorf("all %d retries exhausted: %w", c.MaxRetries+1, lastErr)
		}
	}

	return "", lastErr
}

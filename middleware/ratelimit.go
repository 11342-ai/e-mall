package middleware

import (
	"context"
	"e-mall/repository/cache"
	"e-mall/utils/ctl"
	"e-mall/utils/e"
	"e-mall/utils/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/redis/go-redis/v9"
)

// 滑动窗口 Lua 脚本：原子完成"清理旧记录 + 计数 + 写入 + 设 TTL"
const slidingWindowLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local windowMs = tonumber(ARGV[2])
local maxReqs = tonumber(ARGV[3])

redis.call("ZREMRANGEBYSCORE", key, 0, now - windowMs)

local count = redis.call("ZCARD", key)
if count >= maxReqs then
	return 0
end

redis.call("ZADD", key, now, now)
redis.call("EXPIRE", key, windowMs * 2 / 1000 + 1)
return 1
`

// slidingWindowAllow 原子滑动窗口判断，true=放行 false=拒绝
func slidingWindowAllow(rdb redis.Cmdable, ctx context.Context,
	key string, window time.Duration, maxReqs int) (bool, error) {

	ret, err := rdb.Eval(ctx, slidingWindowLua, []string{key},
		time.Now().UnixMilli(), window.Milliseconds(), maxReqs).Int()
	if err != nil {
		return false, err
	}
	return ret == 1, nil
}

// GlobalRateLimiter 接口级全局限流（令牌桶，本地进程内）
// 注意：多副本部署时每节点独立限流，总 QPS = 单节点 × 副本数
// 如需严格全局限流，应替换为 Redis 中心化令牌桶
func GlobalRateLimiter(rate float64, burst int64) gin.HandlerFunc {
	// bucket 在路由注册时创建一次，闭包捕获，所有请求复用同一个
	bucket := ratelimit.NewBucketWithRate(rate, burst)

	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status": e.ErrorTooManyRequests,
				"msg":    e.GetMsg(e.ErrorTooManyRequests),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GlobalRateLimiterWithBurst 带突发等待的版本（适合非秒杀场景）
func GlobalRateLimiterWithBurst(rate float64, burst int64, maxWait time.Duration) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithRate(rate, burst)
	return func(c *gin.Context) {
		waitTime := bucket.Take(1)
		if waitTime > maxWait {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status": e.ErrorTooManyRequests,
				"msg":    e.GetMsg(e.ErrorTooManyRequests),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserRateLimiter 用户维度限流（基于 Redis 滑动窗口）
// 必须在 AuthMiddleware 之后挂载，否则无法获取用户 ID
func UserRateLimiter(window time.Duration, maxReqs int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, err := ctl.GetUserInfo(c.Request.Context())
		if err != nil || userInfo == nil {
			// 未登录或获取不到用户信息 -> 不拦截（走其他限流层）
			c.Next()
			return
		}

		key := "ratelimit:user:" + string(rune(userInfo.Id)) + ":" + c.FullPath()
		allowed, err := slidingWindowAllow(cache.RedisClient, c.Request.Context(), key, window, maxReqs)
		if err != nil {
			// Redis 异常：打印错误，降级放行
			log.LogrusObj.Errorln("[ratelimit] redis error (user window):", err)
			c.Next()
			return
		}
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status": e.ErrorTooManyRequests,
				"msg":    e.GetMsg(e.ErrorTooManyRequests),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPRateLimiter IP 维度限流（基于 Redis 滑动窗口）
// 辅助防刷手段，NAT 环境下可能误伤，阈值宜宽松
func IPRateLimiter(window time.Duration, maxReqs int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:ip:" + c.ClientIP() + ":" + c.FullPath()
		allowed, err := slidingWindowAllow(cache.RedisClient, c.Request.Context(), key, window, maxReqs)
		if err != nil {
			// Redis 异常：打印错误，降级放行
			log.LogrusObj.Errorln("[ratelimit] redis error (ip window):", err)
			c.Next()
			return
		}
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status": e.ErrorTooManyRequests,
				"msg":    e.GetMsg(e.ErrorTooManyRequests),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

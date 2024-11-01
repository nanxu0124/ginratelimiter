-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象的键
local key = KEYS[1]
-- 窗口大小（毫秒）
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber( ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口的起始时间
local min = now - window

-- 删除窗口范围外的记录
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

-- 获取当前窗口内请求数量
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

-- 如果请求数超过阈值，则返回 true 表示限流
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 否则，添加当前请求并设置过期时间
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return "false"
end
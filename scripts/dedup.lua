const SelectCourseLuaScript = `
-- KEYS[1] : 选课请求 Key (request:studentID:courseID)
-- KEYS[2] : 课程库存 Key (course:stock:courseID)
-- ARGV[1] : 扣减数量 (传 1)

-- 1. 校验选课请求是否存在
local exists = redis.call('get', KEYS[1])
if exists then
    return -1 -- ErrRepeatRequest
end

-- 2. 校验库存并扣减
local stock = redis.call('get', KEYS[2])
if not stock or tonumber(stock) < tonumber(ARGV[1]) then
    return 0 -- ErrStockEmpty
end

redis.call('decrby', KEYS[2], tonumber(ARGV[1]))
redis.call('set',KEYS[1],1,'EX',3600)
return 1 -- 成功
`
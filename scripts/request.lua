local tokens = {}
local file = io.open("tokens_10000.txt", "r")

if file then
    for line in file:lines() do
        table.insert(tokens, line)
    end
    file:close()
    print("Loaded " .. #tokens .. " tokens into memory.")
else
    print("Error: Could not open tokens_10000.txt")
end

-- 每次请求随机分配一个 token
request = function()
    local index = math.random(1, #tokens)
    local token = tokens[index]

    local path = "/auth/select/1"

    local headers = {
        ["Authorization"] = "Bearer " .. token,
        ["Content-Type"] = "application/json"
    }

    return wrk.format("POST", path, headers, nil)
end

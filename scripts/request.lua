local tokens = {}
local file = io.open("tokens_40000.txt", "r")

if not file then
    error("failed to open token file")
end

for line in file:lines() do
    if line ~= "" then
        table.insert(tokens, line)
    end
end
file:close()

local idx = 1

request = function()
    local token = tokens[idx]
    idx = idx + 1
    if idx > #tokens then
        idx = 1
    end

    local headers = {
        ["Authorization"] = "Bearer " .. token,
        ["Content-Type"] = "application/json"
    }

    return wrk.format("POST", "/auth/select/1", headers, nil)
end


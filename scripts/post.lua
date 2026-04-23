local tokens = {}
local token_file = os.getenv("TOKEN_FILE") or "tokens_70000.txt"
local total_threads = tonumber(os.getenv("WRK_THREADS") or "4")
local file = io.open(token_file, "r")

if not file then
    error("failed to open token file: " .. token_file)
end

for line in file:lines() do
    if line ~= "" then
        table.insert(tokens, line)
    end
end
file:close()

local thread_counter = 0

setup = function(thread)
    thread_counter = thread_counter + 1
    thread:set("thread_id", thread_counter)
end

init = function(args)
    idx = thread_id
end

request = function()
    local token = tokens[idx]
    idx = idx + total_threads
    if idx > #tokens then
        idx = thread_id
    end

    local headers = {
        ["Authorization"] = "Bearer " .. token,
        ["Content-Type"] = "application/json"
    }

    return wrk.format("POST", "/auth/select/1", headers, nil)
end

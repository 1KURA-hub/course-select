local token_file = os.getenv("TOKEN_FILE") or "tokens_10000.txt"
local course_id = os.getenv("COURSE_ID") or "1"
local thread_count = tonumber(os.getenv("WRK_THREADS")) or 1

local tokens = {}
local next_thread_id = 0
local thread_id = 0
local issued = 0
local warned = false

function setup(thread)
    thread:set("thread_id", next_thread_id)
    next_thread_id = next_thread_id + 1
end

do
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

    if #tokens == 0 then
        error("token file is empty: " .. token_file)
    end
end

function init(args)
    if thread_count <= 0 then
        error("WRK_THREADS must be greater than 0")
    end
    if thread_count > #tokens then
        error("WRK_THREADS cannot exceed token count")
    end

    issued = 0
end

local function next_token()
    local index = issued * thread_count + thread_id + 1
    if index > #tokens then
        if not warned then
            io.stderr:write("warning: token pool exhausted, requests will start reusing tokens\n")
            warned = true
        end
        index = ((index - 1) % #tokens) + 1
    end

    issued = issued + 1
    return tokens[index]
end

request = function()
    local token = next_token()
    local headers = {
        ["Authorization"] = "Bearer " .. token,
        ["Content-Type"] = "application/json",
    }

    return wrk.format("POST", "/auth/select/" .. course_id, headers, nil)
end

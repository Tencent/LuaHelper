local a = {
    aaa = 1,
    bbb = 2
}

local bbbb = setmetatable({}, {__index = a})

local b = {
    ccc = 1,
    ddd = 2
}

local eeee = setmetatable(b, {})
local ffff = setmetatable(b, {__index = a})
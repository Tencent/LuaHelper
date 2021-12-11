local a1 = function(a, b)
    return {
        a = 1,
        b = 2
    } 
end

local bbbb = setmetatable({}, {__call = a1})
local aaaa = bbbb()

local cccc = setmetatable({}, {__call = function(a, b)
    return {
        a = 1,
        b = 2
    } 
end})

local dddd = cccc()
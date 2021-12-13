test2 = function()
    ---@class test1
    ---@field aa number @AAA
    ---@field cc number @BBB
    ---@type test1
    local default = {
        aa = 1,
        bb = 2,
    }

    return default
end

local yy = test2()
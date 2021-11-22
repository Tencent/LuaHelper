local a = {
    b = 1,
    c = 2,
    f = {
        d = 1,
        g = 1,
    }
}

a.d = {
    a = 1,
    b = 2
}

a.e = 2

local a1 = a
local a2 = a

return a2
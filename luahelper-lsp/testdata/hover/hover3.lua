local a = {
    b = 1,
    c = 2,
    d = {
        a = 1, 
        b = 2
    }
}

a.e = 1
a.f = {
    a = 1,
    b = 2
}

local a1 = a
local a2 = a1
local exa_a = a2
print(exa_a.b)
print(exa_a.e)
print(exa_a.f)
print(exa_a.f.b)
print(exa_a.d)
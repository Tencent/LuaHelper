local function func1()
    return {
        a = 1,
        b = {
            b1 = {
                b2 = 1
            }
        }
    }
end

local c2 = {
    a = 1,
    b = {
        b1 = {
            b2 = 1
        }
    }
}

local f1 = func1()

f1.b.
c2.b.b1.b2
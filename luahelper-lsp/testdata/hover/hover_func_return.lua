function func2()
    return "a", "b", "c"
end

---@param one string @djofo
local function func1(one, two, c)
   if two == 1 then
       return func1()
   end

   print(two.b.c)

   return func2(), func2(), func1
end
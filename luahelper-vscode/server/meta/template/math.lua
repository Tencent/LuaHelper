---@class math @This library provides basic mathematical functions. It provides all its functions and constants inside the table *math*. Functions with the annotation "integer/float" give integer results for integer arguments and float results for non-integer arguments. The rounding functions *math.ceil*, *math.floor*, and *math.modf* return an *integer* when the result fits in the range of an *integer*, or a *float* otherwise.  [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.7)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.7"])
---@field huge       number @The float value HUGE_VAL, a value greater than any other numeric value.
---@field maxinteger integer @An integer with the maximum value for an integer.
---@field mininteger integer @An integer with the minimum value for an integer.
---@field pi         number @The value of π.
math = {}

--- Returns the absolute value of `x`. (integer/float)
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.abs)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.abs"])
function math.abs(x) end

--- Returns the arc cosine of `x` (in radians).
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.acos)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.acos"])
function math.acos(x) end

--- Returns the arc sine of `x` (in radians).
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.asin)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.asin"])
function math.asin(x) end

--- Returns the arc tangent of `y/x` (in radians), but uses the signs of both
--- parameters to find the quadrant of the result. (It also handles correctly
--- the case of `x` being zero.)
---
--- The default value for `x` is 1, so that the call `math.atan(y)`` returns the
--- arc tangent of `y`.
---@param y  number
---@param x? number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.atan)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.atan"])
function math.atan(y, x) end

-- Returns the arc tangent of y/x (in radians), but uses the signs of both parameters to find the quadrant of the result. (It also handles correctly the case of x being zero.)
---@version lua<5.2
---@param y number
---@param x number
---@return number
function math.atan2(y, x) end

--- Returns the smallest integer larger than or equal to `x`.
---@param x number
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.ceil)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.ceil"])
function math.ceil(x) end

--- Returns the cosine of `x` (assumed to be in radians).
---@param x number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.cos)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.cos"])
function math.cos(x) end

-- Returns the hyperbolic cosine of x.
---@version <lua5.2
---@param x number
---@return number
function math.cosh(x) end

--- Converts the angle `x` from radians to degrees.
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.deg)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.deg"])
function math.deg(x) end

--- Returns the value *e^x* (where e is the base of natural logarithms).
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.exp)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.exp"])
function math.exp(x) end

--- Returns the largest integer smaller than or equal to `x`.
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.abs)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.abs"])
function math.floor(x) end

--- Returns the remainder of the division of `x` by `y` that rounds the quotient towards zero. (integer/float)
---@param x number
---@param y number
---@return number
function math.fmod(x, y) end

-- Returns m and e such that x = m2e, e is an integer and the absolute value of m is in the range [0.5, 1) (or zero when x is zero).
---@version <lua5.2
---@param x number
---@return number m
---@return number e
function math.frexp(x) end

---@version lua<5.2
---@param m number
---@param e number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.abs)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.abs"])
function math.ldexp(m, e) end

--- Returns the logarithm of `x` in the given base. The default for `base` is
--- *e* (so that the function returns the natural logarithm of `x`).
---@param x     number
---@param base? integer
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.log)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.log"])
function math.log(x, base) end

--- Returns the argument with the maximum value, according to the Lua operator
--- `<`. (integer/float)
---@param x number
---@vararg number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.max)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.max"])
function math.max(x, ...) end

--- Returns the argument with the minimum value, according to the Lua operator
--- `<`. (integer/float)
---@param x number
---@vararg number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.min)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.min"])
function math.min(x, ...) end

--- Returns the integral part of `x` and the fractional part of `x`. Its second
--- result is always a float.
---@param x number
---@return integer
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.modf)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.modf"])
function math.modf(x) end

--Returns xy. (You can also use the expression x^y to compute this value.)
---@version <lua5.2
---@param x number
---@param y number
---@return number
function math.pow(x, y) end

--- Converts the angle `x` from degrees to radians.'
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.rad)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.rad"])
function math.rad(x) end

--- When called without arguments, returns a pseudo-random float with uniform
--- distribution in the range *[0,1)*. When called with two integers `m` and
--- `n`, `math.random` returns a pseudo-random integer with uniform distribution
--- in the range *[m, n]*. The call `math.random(n)` is equivalent to `math
--- .random`(1,n).
---@overload fun():number
---@overload fun(m: integer):integer
---@param m integer
---@param n integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.random)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.random"])
function math.random(m, n) end


--When called with at least one argument, the integer parameters x and y are joined into a 128-bit seed that is used to reinitialize the pseudo-random generator; equal seeds produce equal sequences of numbers. The default for y is zero.
--
--When called with no arguments, Lua generates a seed with a weak attempt for randomness.
--
--This function returns the two seed components that were effectively used, so that setting them again repeats the sequence.
--
--To ensure a required level of randomness to the initial state (or contrarily, to have a deterministic sequence, for instance when debugging a program), you should call `math.randomseed` with explicit arguments.
---@param x? integer
---@param y? integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.randomseed)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.randomseed"])
function math.randomseed(x, y) end


--- Returns the sine of `x` (assumed to be in radians).
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.sin)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.sin"])
function math.sin(x) end

-- Returns the hyperbolic sine of x.
---@version <lua5.2
---@param x number
---@return number
function math.sinh(x) end

--- Returns the square root of `x`. (You can also use the expression `x^0.5` to
--- compute this value.)
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.sqrt)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.sqrt"])
function math.sqrt(x) end

--- Returns the tangent of `x` (assumed to be in radians).
---@param x number
---@return number
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.tan)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.tan"])
function math.tan(x) end

--Returns the hyperbolic tangent of x.
---@version <lua5.2
---@param x number
---@return number
function math.tanh(x) end

--- If the value `x` is convertible to an *integer*, returns that *integer*.
--- Otherwise, returns `fail`.
---@param x number
---@return integer?
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.tointeger)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.tointeger"])
function math.tointeger(x) end

--- Returns "`integer`" if `x` is an integer, "`float`" if it is a float, or
--- **nil** if `x` is not a number.
---@param x any
---@return string @"integer" or "float" or "nil"
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.type)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.type"])
function math.type(x) end

--- Returns a boolean, true if and only if integer `m` is below integer `n` when
--- they are compared as unsigned integers.
---@param m integer
---@param n integer
---@return boolean
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-math.ult)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-math.ult"])
function math.ult(m, n) end

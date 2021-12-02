---@version lua5.2
---@class bit32lib @This library provides bitwise operations. It provides all its functions inside the table bit32. [`View online doc`](https://www.lua.org/manual/5.2/manual.html#6.7)  
bit32 = {}

---Returns the number x shifted disp bits to the right. The number disp may be any representable integer. Negative displacements shift to the left.
---
---This shift operation is what is called arithmetic shift. Vacant bits on the left are filled with copies of the higher bit of x; vacant bits on the right are filled with zeros. In particular, displacements with absolute values higher than 31 result in zero or 0xFFFFFFFF (all original bits are shifted out).
---@param x    integer
---@param disp integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.arshift)  
function bit32.arshift(x, disp) end

---Returns the bitwise and of its operands.
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.band)  
function bit32.band(...) end

--Returns the bitwise negation of x. For any integer x, the following identity holds:
--
--   *assert(bit32.bnot(x) == (-1 - x) % 2^32)*
---@param x integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.bnot)  
function bit32.bnot(x) end

--Returns the bitwise or of its operands.
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.bor)  
function bit32.bor(...) end

---Returns a boolean signaling whether the bitwise and of its operands is different from zero.
---@return boolean
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.btest)  
function bit32.btest(...) end

---Returns the bitwise exclusive or of its operands.
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.bxor)  
function bit32.bxor(...) end

--Returns the unsigned number formed by the bits field to field + width - 1 from n. Bits are numbered from 0 (least significant) to 31 (most significant). All accessed bits must be in the range [0, 31].
--
--The default for width is 1.
---@param n      integer
---@param field  integer
---@param width? integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.extract)  
function bit32.extract(n, field, width) end

--Returns a copy of n with the bits field to field + width - 1 replaced by the value v. See bit32.extract for details about field and width.
---@param n integer
---@param v integer
---@param field  integer
---@param width? integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.replace)  
function bit32.replace(n, v, field, width) end

--Returns the number x rotated disp bits to the left. The number disp may be any representable integer.
--
--For any valid displacement, the following identity holds:
--
--     assert(bit32.lrotate(x, disp) == bit32.lrotate(x, disp % 32))
--In particular, negative displacements rotate to the right.
---@param x     integer
---@param distp integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.lrotate)  
function bit32.lrotate(x, distp) end

---Returns the number x shifted disp bits to the left. The number disp may be any representable integer. Negative displacements shift to the right. In any direction, vacant bits are filled with zeros. In particular, displacements with absolute values higher than 31 result in zero (all bits are shifted out).
--
--For positive displacements, the following equality holds:
--
-- assert(bit32.lshift(b, disp) == (b * 2^disp) % 2^32)
---@param x     integer
---@param distp integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.lshift)  
function bit32.lshift(x, distp) end

---Returns the number x rotated disp bits to the right. The number disp may be any representable integer.
--
--For any valid displacement, the following identity holds:
--
--assert(bit32.rrotate(x, disp) == bit32.rrotate(x, disp % 32))
--In particular, negative displacements rotate to the left.
---@param x     integer
---@param distp integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.rrotate)  
function bit32.rrotate(x, distp) end

---Returns the number x shifted disp bits to the right. The number disp may be any representable integer. Negative displacements shift to the left. In any direction, vacant bits are filled with zeros. In particular, displacements with absolute values higher than 31 result in zero (all bits are shifted out).
---
---For positive displacements, the following equality holds:
--
-- *assert(bit32.rshift(b, disp) == math.floor(b % 2^32 / 2^disp))*
--This shift operation is what is called logical shift.
---@param x     integer
---@param distp integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.2/manual.html#pdf-bit32.rshift)  
function bit32.rshift(x, distp) end

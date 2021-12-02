---@class one11
---@field astring string
---@field bnumber number
---@field cany any
---@field dfun fun(a:number, b:number)

---@type one11
local aaaa1 = {}
aaaa1.abc = 1

local aaaaa2 = aaaa1

aaaa1.bnumber = 2
aaaa1.astring = "sfs"
aaaaa2.bnumber = 2
aaaaa2.astring = "sfs"
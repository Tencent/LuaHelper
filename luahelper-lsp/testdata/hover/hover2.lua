local req_v = require("hover2_requrie")

print(req_v.b)
print(req_v.e)
print(req_v.f)
print(req_v.f.d)
print(req_v.d)
print(req_v.d.b)

local req_f = require("hover2_requrie").f
print(req_f.d)

local req_d = require("hover2_requrie").d
print(req_d.a)


-- hover2_requrie2
local req_v = require("hover2_requrie2")

print(req_v.b)
print(req_v.f)
print(req_v.f.d)
print(req_v.d)

local req_f = require("hover2_requrie").f
print(req_f.d)

print(require("hover2_requrie").d.a)
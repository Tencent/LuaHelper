---@class utf8 @This library provides basic support for UTF-8 encoding. It provides all its functions inside the table utf8. This library does not provide any support for Unicode other than the handling of the encoding. Any operation that needs the meaning of a character, such as character classification, is outside its scope.  [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.5)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.5"])
---@field charpattern string
utf8 = {}

---@type string @The pattern (a string, not a function) "`[\0-\x7F\xC2-\xF4][\x80-\xBF]*`", which matches exactly one UTF-8 byte sequence, assuming that the subject is a valid UTF-8 string. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.charpattern)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.charpattern"])
utf8.charpattern = ""


--- Receives zero or more integers, converts each one to its corresponding
--- UTF-8 byte sequence and returns a string with the concatenation of all
--- these sequences.
---@param code integer
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.char)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.char"])
function utf8.char(code, ...) end

--Returns values so that the construction
--
--  *for p, c in utf8.codes(s) do body end*
-- will iterate over all UTF-8 characters in string `s`, with p being the position (in bytes) and `c` the code point of each character. It raises an error if it meets any invalid byte sequence.
---@param s    string
---@param lax? boolean
---@return fun():integer, integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.codes)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.codes"])
function utf8.codes(s, lax) end

--- Returns the codepoints (as integers) from all characters in `s` that start
--- between byte position `i` and `j` (both included). The default for `i` is
--- 1  and for `j` is `i`. It raises an error if it meets any invalid byte
--- sequence.
---@param s    string
---@param i?   integer
---@param j?   integer
---@param lax? boolean
---@return integer code
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.codepoint)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.codepoint"])
function utf8.codepoint(s, i, j, lax) end

--- Returns the number of UTF-8 characters in string `s` that start between
--- positions `i` and `j` (both inclusive). The default for `i` is 1 and for
--- `j` is -1. If it finds any invalid byte sequence, returns a false value
--- plus the position of the first invalid byte.
---@param s    string
---@param i?   integer
---@param j?   integer
---@param lax? boolean
---@return integer?
---@return integer? errpos
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.len)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.len"])
function utf8.len(s, i, j, lax) end

--- Returns the position (in bytes) where the encoding of the `n`-th character
--- of `s` (counting from position `i`) starts. A negative `n` gets
--- characters before position `i`. The default for `i` is 1 when `n` is
--- non-negative and `#s + 1` otherwise, so that `utf8.offset(s, -n)` gets the
--- offset of the `n`-th character from the end of the string. If the
--- specified character is neither in the subject nor right after its end,
--- the function returns nil. As a special case, when `n` is 0 the function
--- returns the start of the encoding of the character that contains the `i`-th
--- byte of `s`.
---
--- This function assumes that `s` is a valid UTF-8 string.
---@param s string
---@param n integer
---@param i integer
---@return integer p
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-utf8.offset)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-utf8.offset"])
function utf8.offset(s, n, i) end
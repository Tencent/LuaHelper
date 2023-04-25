---@class stringlib @The type *string* represents immutable sequences of bytes. Lua is 8-bit clean: strings can contain any 8-bit value, including embedded zeros('`\0`'). Lua is also encoding-agnostic; it makes no assumptions about the contents of a string. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.4)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.4"])
string = {}

--- Returns the internal numerical codes of the characters `s[i]`, `s[i+1]`,
--- ..., `s[j]`. The default value for `i` is 1; the default value for `j`
--- is `i`. These indices are corrected following the same rules of function
--- `string.sub`.
---
--- Note that numerical codes are not necessarily portable across platforms.
---@param s  string
---@param i? integer
---@param j? integer
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.byte)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.byte"])
function string.byte(s, i, j) end


--- Receives zero or more integers. Returns a string with length equal to
--- the number of arguments, in which each character has the internal numerical
--- code equal to its corresponding argument.
---
--- Note that numerical codes are not necessarily portable across platforms.
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.char)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.char"])
function string.char(byte, ...) end

--- Returns a string containing a binary representation (*a binary chunk*) of
--- the given function, so that a later `load` on this string returns a
--- copy of the function (but with new upvalues). If strip is a true value, the
--- binary representation may not include all debug information about the
--- function, to save space.
---
--- Functions with upvalues have only their number of upvalues saved. When (re)
--- loaded, those upvalues receive fresh instances containing **nil**. (You can
--- use the debug library to serialize and reload the upvalues of a function in
--- a way adequate to your needs.)
---@param f      function
---@param strip? boolean
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.dump)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.dump"])
function string.dump(f, strip) end

--- Looks for the first match of `pattern` in the string `s`. If it finds a
--- match, then `find` returns the indices of `s` where this occurrence starts
--- and ends; otherwise, it returns **nil**. A third, optional numerical
--- argument `init` specifies where to start the search; its default value is 1
--- and can be negative. A value of **true** as a fourth, optional argument
--- `plain` turns off the pattern matching facilities, so the function does a
--- plain "find substring" operation, with no characters in `pattern` being
--- considered "magic". Note that if `plain` is given, then `init` must be given
--- as well.
---
--- If the pattern has captures, then in a successful match the captured values
--- are also returned, after the two indices.
---@param s       string
---@param pattern string
---@param init?   integer
---@param plain?  boolean
---@return integer start
---@return integer end
---@return ... captured
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.find)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.find"])
function string.find(s, pattern, init, plain) end

--- Returns a formatted version of its variable number of arguments following
--- the description given in its first argument (which must be a string). The
--- format string follows the same rules as the ISO C function `sprintf`. The
--- only differences are that the options/modifiers `*`, `h`, `L`, `l`, `n`, and
--- `p` are not supported and that there is an extra option, `q`.
---@param s string
---@vararg string
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.format)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.format"])
function string.format(s, ...) end

--- Returns an iterator function that, each time it is called, returns the
--- next captures from `pattern` over the string `s`. If `pattern` specifies no
--- captures, then the whole match is produced in each call.
---
--- As an example, the following loop will iterate over all the words from
--- string `s`, printing one per line:
---
--- `s = "hello world from Lua"`
--- `for w in string.gmatch(s, "%a+") do`
---  > `print(w)`
--- `end`
---
--- The next example collects all pairs `key=value` from the given string into a
--- table:
---
--- `t = {}`
---  s = "from=world, to=Lua"`
--- `for k, v in string.gmatch(s, "(%w+)=(%w+)") do`
---  > `t[k] = v`
--- `end`
---
--- For this function, a caret '`^`' at the start of a pattern does not work as
--- an anchor, as this would prevent the iteration.
---@param s       string
---@param pattern string
---@param init?   integer
---@return fun():string, table
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.gmatch)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.gmatch"])
function string.gmatch(s, pattern, init) end

--- Returns a copy of `s` in which all (or the first `n`, if given)
--- occurrences of the `pattern` have been replaced by a replacement string
--- specified by `repl`, which can be a string, a table, or a function. `gsub`
--- also returns, as its second value, the total number of matches that
--- occurred.
---
--- If `repl` is a string, then its value is used for replacement. The character
--- `%` works as an escape character: any sequence in `repl` of the form `%n`,
--- with *n* between 1 and 9, stands for the value of the *n*-th captured
--- substring (see below). The sequence `%0` stands for the whole match. The
--- sequence `%%` stands for a single `%`.
---
--- If `repl` is a table, then the table is queried for every match, using
--- the first capture as the key; if the pattern specifies no captures, then
--- the whole match is used as the key.
---
--- If `repl` is a function, then this function is called every time a match
--- occurs, with all captured substrings passed as arguments, in order; if
--- the pattern specifies no captures, then the whole match is passed as a
--- sole argument.
---
--- If the value returned by the table query or by the function call is a
--- string or a number, then it is used as the replacement string; otherwise,
--- if it is false or nil, then there is no replacement (that is, the original
--- match is kept in the string).
---@param s       string
---@param pattern string
---@param repl    string|table|function
---@param n       integer
---@return string
---@return integer count
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.gsub)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.gsub"])
function string.gsub(s, pattern, repl, n) end

--- Receives a string and returns its length. The empty string `""` has
--- length 0. Embedded zeros are counted, so `"a\000bc\000"` has length 5.
---@param s string
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.len)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.len"])
function string.len(s) end

--- Receives a string and returns a copy of this string with all uppercase
--- letters changed to lowercase. All other characters are left unchanged. The
--- definition of what an uppercase letter is depends on the current locale.
---@param s string
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.lower)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.lower"])
function string.lower(s) end

--- Looks for the first *match* of `pattern` in the string `s`. If it
--- finds one, then `match` returns the captures from the pattern; otherwise
--- it returns **nil**. If `pattern` specifies no captures, then the whole match
--- is returned. A third, optional numerical argument `init` specifies where
--- to start the search; its default value is 1 and can be negative.
---@param s       string
---@param pattern string
---@param init?   integer
---@return string captured
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.match)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.match"])
function string.match(s, pattern, init) end

--- Returns a binary string containing the values `v1`, `v2`, etc. packed (that
--- is, serialized in binary form) according to the format string `fmt`.
---@version >lua5.3
---@param fmt string
---@param v1  string
---@param v2? string
---@vararg string
---@return string binary
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.pack)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.pack"])
function string.pack(fmt, v1, v2, ...) end

--- Returns the size of a string resulting from `string.pack` with the given
--- format. The format string cannot have the variable-length options '`s`' or '`z`'
---@version >lua5.3
---@param fmt string
---@return integer
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.packsize)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.packsize"])
function string.packsize(fmt) end

--- Returns a string that is the concatenation of `n` copies of the string
--- `s` separated by the string `sep`. The default value for `sep` is the empty
--- string (that is, no separator). Returns the empty string if n is not
--- positive.
---
--- Note that it is very easy to exhaust the memory of your machine with a
--- single call to this function.
---@param s    string
---@param n    integer
---@param sep? string
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.rep)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.rep"])
function string.rep(s, n, sep) end

--- Returns a string that is the string `s` reversed.
---@param s string
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.reverse)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.reverse"])
function string.reverse(s) end

--- Returns the substring of `s` that starts at `i` and continues until
--- `j`; `i` and `j` can be negative. If `j` is absent, then it is assumed to
--- be equal to -1 (which is the same as the string length). In particular,
--- the call `string.sub(s,1,j)` returns a prefix of `s` with length `j`, and
--- `string.sub(s, -i)` (for a positive i) returns a suffix of `s` with length
--- `i`.
---
--- If, after the translation of negative indices, `i` is less than 1, it is
--- corrected to 1. If `j` is greater than the string length, it is corrected to
--- that length. If, after these corrections, `i` is greater than `j`, the
--- function returns the empty string.
---@param s  string
---@param i  integer
---@param j? integer
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.sub)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.sub"])
function string.sub(s, i, j) end

--- Returns the values packed in string `s` according to the format string
--- `fmt`. An optional `pos` marks where to start reading in `s` (default is 1).
--- After the read values, this function also returns the index of the first
--- unread byte in `s`.
---@version >lua5.3
---@param fmt  string
---@param s    string
---@param pos? integer
---@return any
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.unpack)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.unpack"])
function string.unpack(fmt, s, pos) end

--- Receives a string and returns a copy of this string with all lowercase
--- letters changed to uppercase. All other characters are left unchanged. The
--- definition of what a lowercase letter is depends on the current locale.
---@param s string
---@return string
--[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-string.upper)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-string.upper"])
function string.upper(s) end

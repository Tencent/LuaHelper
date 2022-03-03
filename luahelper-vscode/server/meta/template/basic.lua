---@type table
arg = {}

-- _ENV is the global environment table.
_ENV = {}

--- Calls error if the value of its argument `v` is false (i.e., **nil** or **false**); otherwise, returns all its arguments. In case of error, `message` is the error object; when absent, it defaults to "assertion failed!"
---@param v any
---@param message? string
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-assert)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-assert"])
function assert(v, message) end

---@alias cgopt
---| '"collect"'      # performs a full garbage-collection cycle. This is the default option.
---| '"stop"'         # stops automatic execution of the garbage collector.
---| '"restart"'      # restarts automatic execution of the garbage collector.
---| '"count"'        # returns the total memory in use by Lua in Kbytes.
---| '"step"'         # Runs one step of garbage collection. The larger the second argument is, the larger this step will be. The collectgarbage will return true if the triggered step was the last step of a garbage-collection cycle.
---| '"isrunning"'    # returns a boolean that tells whether the collector is running (i.e., not stopped).
---| '"incremental"'  # Change the collector mode to incremental.
---| '"generational"' # Change the collector mode to generational. This option can be followed by two numbers: the garbage-collector minor multiplier and the major multiplier.
---| '"setpause"'     # sets `arg` as the new value for the *pause* of the collector Returns the previous value for *pause`.
---| '"setstepmul"'   # Sets the value given as second parameter divided by 100 to the garbage step multiplier variable. Its uses are as discussed a little above.

---
--- This function is a generic interface to the garbage collector. It performs
--- different functions according to its first argument, `opt`:
---
--- **"collect"**: performs a full garbage-collection cycle. This is the default
--- option.
--- **"stop"**: stops automatic execution of the garbage collector. The
--- collector will run only when explicitly invoked, until a call to restart it.
--- **"restart"**: restarts automatic execution of the garbage collector.
--- **"count"**: returns the total memory in use by Lua in Kbytes. The value has
--- a fractional part, so that it multiplied by 1024 gives the exact number of
--- bytes in use by Lua (except for overflows).
--- **"step"**: performs a garbage-collection step. The step "size" is
--- controlled by `arg`. With a zero value, the collector will perform one basic
--- (indivisible) step. For non-zero values, the collector will perform as if
--- that amount of memory (in KBytes) had been allocated by Lua. Returns
--- **true** if the step finished a collection cycle.
--- **"setpause"**: sets `arg` as the new value for the *pause* of the collector
--- Returns the previous value for *pause`.
--- **"incremental"**: Change the collector mode to incremental. This option can
--- be followed by three numbers: the garbage-collector pause, the step
--- multiplier, and the step size.
--- **"generational"**: Change the collector mode to generational. This option
--- can be followed by two numbers: the garbage-collector minor multiplier and
--- the major multiplier.
--- **"isrunning"**: returns a boolean that tells whether the collector is
--- running (i.e., not stopped).
---@param opt? cgopt
---@param arg? string
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-collectgarbage)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-collectgarbage"])
function collectgarbage(opt, arg) end

--- Opens the named file and executes its contents as a Lua chunk. When called
--- without arguments, `dofile` executes the contents of the standard input
--- (`stdin`). Returns all values returned by the chunk. In case of errors,
--- `dofile` propagates the error to its caller (that is, `dofile` does not run
--- in protected mode).
---@param filename? string
---@return table
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-dofile)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-dofile"])
function dofile(filename) end

--- Terminates the last protected function called and returns `message` as the
--- error object. Function `error` never returns. Usually, `error` adds some
--- information about the error position at the beginning of the message, if the
--- message is a string. The `level` argument specifies how to get the error
--- position. With level 1 (the default), the error position is where the
--- `error` function was called. Level 2 points the error to where the function
--- that called `error` was called; and so on. Passing a level 0 avoids the
--- addition of error position information to the message.
---@param message string
---@param level? number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-error)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-error"])
function error(message, level) end

---@class _G @A global variable (not a function) that holds the global environment. Lua itself does not use this variable; changing its value does not affect any environment, nor vice versa. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-_G)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-_G"])
_G = {}

---Returns the current environment in use by the function. *f* can be a Lua function or a number that specifies the function at that stack level: Level 1 is the function calling `getfenv`. If the given function is not a Lua function, or if f is 0, `getfenv` returns the global environment. The default for *f* is 1.
---@version lua5.1
---@param f? function
---@return table
function getfenv(f) end

--- If `object` does not have a metatable, returns **nil**. Otherwise, if the
--- object's metatable has a `"__metatable"` field, returns the associated
--- value. Otherwise, returns the metatable of the given object.
---@param object any
---@return table metatable
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-getmetatable)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-getmetatable"])
function getmetatable(object) end

--- Returns three values (an iterator function, the table `t`, and 0) so that the construction
--- `for i,v in ipairs(t) do` *body* `end`
--- will iterate over the key–value pairs (1,`t[1]`), (2,`t[2]`), ..., up to the first absent index.
---@generic V
---@param t table<number, V>|V[]
---@return fun(tbl: table<number, V>):number, V
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-ipairs)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-ipairs"])
function ipairs(t) end

---@alias loadmode
---| '"b"'  # ---#DESTAIL 'loadmode.b'
---| '"t"'  # ---#DESTAIL 'loadmode.t'
---| '"bt"' # ---#DESTAIL 'loadmode.bt'

--- Loads a chunk.
--- If `chunk` is a string, the chunk is this string. If `chunk` is a function,
--- `load` calls it repeatedly to get the chunk pieces. Each call to `chunk`
--- must return a string that concatenates with previous results. A return of
--- an empty string, **nil**, or no value signals the end of the chunk.
---
--- If there are no syntactic errors, returns the compiled chunk as a function;
--- otherwise, returns **nil** plus the error message.
---
--- If the resulting function has upvalues, the first upvalue is set to the
--- value of `env`, if that parameter is given, or to the value of the global
--- environment. Other upvalues are initialized with **nil**. (When you load a
--- main chunk, the resulting function will always have exactly one upvalue, the
--- _ENV variable. However, when you load a binary chunk created from a
--- function (see string.dump), the resulting function can have an arbitrary
--- number of upvalues.) All upvalues are fresh, that is, they are not shared
--- with any other function.
---
--- `chunkname` is used as the name of the chunk for error messages and debug
--- information. When absent, it defaults to `chunk`, if `chunk` is a string,
--- or to "=(`load`)" otherwise.
---
--- The string `mode` controls whether the chunk can be text or binary (that is,
--- a precompiled chunk). It may be the string "b" (only binary chunks), "t"
--- (only text chunks), or "bt" (both binary and text). The default is "bt".
---
--- Lua does not check the consistency of binary chunks. Maliciously crafted
--- binary chunks can crash the interpreter.
---@param chunk fun():string
---@param chunkname? string
---@param mode? loadmode
---@param env? any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-load)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-load"])
function load(chunk, chunkname, mode, env) end

--- Similar to `load`, but gets the chunk from file `filename` or from the standard input, if no file name is given.
---@param filename? string
---@param mode? loadmode
---@param env? any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-loadfile)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-loadfile"])
function loadfile(filename, mode, env) end

-- Similar to `load`, but gets the chunk from the given string. To load and run a given string, use the idiom assert(loadstring(s))() When absent, chunkname defaults to the given string.
---@version lua5.1
---@param text       string
---@param chunkname? string
---@return function
---@return string error_message
function loadstring(text, chunkname) end

-- Creates a `module`. If there is a table in package.loaded[name], this table is the `module`. Otherwise, if there is a global table t with the given name, this table is the module. Otherwise creates a new table t and sets it as the value of the global name and the value of package.loaded[name]. This function also initializes t._NAME with the given name, t._M with the module (t itself), and t._PACKAGE with the package name (the full module name minus last component; see below). Finally, module sets t as the new environment of the current function and the new value of package.loaded[name], so that *require* returns t.
---@version lua5.1
---@param name string
function module(name, ...) end

--- Allows a program to traverse all fields of a table. Its first argument is
--- a table and its second argument is an index in this table. `next` returns
--- the next index of the table and its associated value. When called with
--- **nil** as its second argument, `next` returns an initial index and its
--- associated value. When called with the last index, or with **nil** in an
--- empty table, `next` returns **nil**. If the second argument is absent, then
--- it is interpreted as **nil**. In particular, you can use `next(t)` to check
--- whether a table is empty.
---
--- The order in which the indices are enumerated is not specified, *even for
--- numeric indices*. (To traverse a table in numerical order, use a numerical
--- **for**.)
---
--- The behavior of `next` is undefined if, during the traversal, you assign
--- any value to a non-existent field in the table. You may however modify
--- existing fields. In particular, you may set existing fields to nil.
---@param table table
---@param index? any
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-next)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-next"])
function next(table, index) end

--- If `t` has a metamethod `__pairs`, calls it with `t` as argument and returns the first three results from the call.
---
--- Otherwise, returns three values: the `next` function, the table `t`, and
--- **nil**, so that the construction
--- `for k,v in pairs(t) do *body* end`
--- will iterate over all key–value pairs of table `t`.
---
--- See function `next` for the caveats of modifying the table during its traversal.
---@generic K, V
---@param t table<K, V>|V[]
---@return fun(tbl: table<K, V>):K, V
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-pairs)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-pairs"])
function pairs(t) end

--- Calls function `f` with the given arguments in *protected mode*. This
--- means that any error inside `f` is not propagated; instead, `pcall` catches
--- the error and returns a status code. Its first result is the status code (a
--- boolean), which is true if the call succeeds without errors. In such case,
--- `pcall` also returns all results from the call, after this first result. In
--- case of any error, `pcall` returns **false** plus the error message.
---@param f fun():any
---@param arg1 ? table
---@return boolean|table
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-pcall)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-pcall"])
function pcall(f, arg1, ...) end

--- Receives any number of arguments, and prints their values to `stdout`, using the `tostring` function to convert them to strings. `print` is not intended for formatted output, but only as a quick way to show a value, for instance for debugging. For complete control over the output, use `string.format` and `io.write`.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-print)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-print"])
function print(...) end

--- Checks whether `v1` is equal to `v2`, without the `__eq` metamethod. Returns a boolean.
---@param v1 any
---@param v2 any
---@return boolean
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-rawequal)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-rawequal"])
function rawequal(v1, v2) end

--- Gets the real value of `table[index]`, the `__index` metamethod. `table` must be a table; `index` may be any value.
---@param table table
---@param index any
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-rawget)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-rawget"])
function rawget(table, index) end

--- Returns the length of the object `v`, which must be a table or a string, without invoking any metamethod. Returns an integer number.
---@param v string|table
---@return number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-rawlen)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-rawlen"])
function rawlen(v) end

--- Sets the real value of `table[index]` to `value`, without invoking the `__newindex` metamethod. `table` must be a table, `index` any value different from **nil** and NaN, and `value` any Lua value.
---@param table table
---@param index any
---@param value any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-rawset)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-rawset"])
function rawset(table, index, value) end


--- Loads the given module. The function starts by looking into the
--- 'package.loaded' table to determine whether `modname` is already
--- loaded. If it is, then `require` returns the value stored at
--- `package.loaded[modname]`. Otherwise, it tries to find a *loader* for
--- the module.
---
--- To find a loader, `require` is guided by the `package.searchers` sequence.
--- By changing this sequence, we can change how `require` looks for a module.
--- The following explanation is based on the default configuration for
--- `package.searchers`.
---
--- First `require` queries `package.preload[modname]`. If it has a value,
--- this value (which should be a function) is the loader. Otherwise `require`
--- searches for a Lua loader using the path stored in `package.path`. If
--- that also fails, it searches for a C loader using the path stored in
--- `package.cpath`. If that also fails, it tries an *all-in-one* loader (see
--- `package.loaders`).
---
--- Once a loader is found, `require` calls the loader with a two argument:
--- `modname` and an extra value dependent on how it got the loader. (If the
--- loader came from a file, this extra value is the file name.) If the loader
--- returns any non-nil value, require assigns the returned value to
--- `package.loaded[modname]`. If the loader does not return a non-nil value and
--- has not assigned any value to `package.loaded[modname]`, then `require`
--- assigns true to this entry. In any case, require returns the final value of
--- `package.loaded[modname]`.
---
--- If there is any error loading or running the module, or if it cannot find
--- any loader for the module, then `require` raises an error.
---@param modname string
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-require)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-require"])
function require(modname) end

--- If `index` is a number, returns all arguments after argument number
--- `index`. a negative number indexes from the end (-1 is the last argument).
--- Otherwise, `index` must be the string "#", and `select` returns
--- the total number of extra arguments it received.
---@generic T
---@param index number|string
---@vararg T
---@return T
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-select)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-select"])
function select(index, ...) end


--- Sets the environment to be used by the given function. f can be a Lua function or a number that specifies the function at that stack level: Level 1 is the function calling `setfenv`.`setfenv` returns the given function. As a special case, when f is 0 `setfenv` changes the environment of the running thread. In this case, `setfenv`  returns no values.
---@version lua5.1
---@param f     function|integer
---@param table table
---@return function
function setfenv(f, table) end

--- Sets the metatable for the given table. (To change the metatable of other
--- types from Lua code, you must use the debug library.) If `metatable`
--- is **nil**, removes the metatable of the given table. If the original
--- metatable has a `"__metatable"` field, raises an error.
---
--- This function returns `table`.
---@generic T
---@param table T
---@param metatable table
---@return T
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-setmetatable)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-setmetatable"])
function setmetatable(table, metatable) end

--- When called with no `base`, `tonumber` tries to convert its argument to a
--- number. If the argument is already a number or a string convertible to a
--- number, then `tonumber` returns this number; otherwise, it returns **nil**.
---
--- The conversion of strings can result in integers or floats, according to the
--- lexical conventions of Lua. (The string may have leading and trailing
--- spaces and a sign.)
---
--- When called with `base`, then e must be a string to be interpreted as an
--- integer numeral in that base. The base may be any integer between 2 and 36,
--- inclusive. In bases above 10, the letter 'A' (in either upper or lower case)
--- represents 10, 'B' represents 11, and so forth, with 'Z' representing 35. If
--- the string `e` is not a valid numeral in the given base, the function
--- returns **nil**.
---@param e string
---@param base? number
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-tonumber)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-tonumber"])
function tonumber(e, base) end

--- Receives a value of any type and converts it to a string in a human-readable
--- format. (For complete control of how numbers are converted, use `string
--- .format`).
---
--- If the metatable of `v` has a `__tostring` field, then `tostring` calls
--- the corresponding value with `v` as argument, and uses the result of the
--- call as its result.
---@param v any
---@return string
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-tostring)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-tostring"])
function tostring(v) end

---@alias typestr
---| '"nil"'
---| '"number"'
---| '"string"'
---| '"boolean"'
---| '"table"'
---| '"function"'
---| '"thread"'
---| '"userdata"'

--- Returns the type of its only argument, coded as a string. The possible
--- results of this function are "`nil`" (a string, not the value **nil**),
--- "`number`", "`string`", "`boolean`", "`table`", "`function`", "`thread`",
--- and "`userdata`".
---@param v any
---@return string
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-type)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-type"])
function type(v) end

_VERSION = 'Lua 5.4'


-- Emits a warning with a message composed by the concatenation of all its arguments (which should be strings).
---
--- By convention, a one-piece message starting with '@' is intended to be a control message, which is a message to the warning system itself. In particular, the standard warning function in Lua recognizes the control messages "@off", to stop the emission of warnings, and "@on", to (re)start the emission; it ignores unknown control messages.
---@version lua5.4
---@param message string
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-warn)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-warn"])
function warn(message, ...) end

--- This function is similar to `pcall`, except that it sets a new message handler `msgh`.
---@param f fun():any
---@param msgh fun():string
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-xpcall)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-xpcall"])
function xpcall(f, msgh, arg1, ...) end

--- Returns the elements from the given table. This function is equivalent to
--   ```return list[i], list[i+1], ..., list[j]```
--   except that the above code can be written only for a fixed number of elements. By default, *i* is 1 and *j* is the length of the list, as defined by the length operator 
---@version lua5.1
---@param list table
---@param i?   integer
---@param j?   integer
function unpack(list, i, j) end

--- Loads the given module. The function starts by looking into the
--- 'package.loaded' table to determine whether `modname` is already
--- loaded. If it is, then `require` returns the value stored at
--- `package.loaded[modname]`. Otherwise, it tries to find a *loader* for
--- the module.
---
--- To find a loader, `require` is guided by the `package.searchers` sequence.
--- By changing this sequence, we can change how `require` looks for a module.
--- The following explanation is based on the default configuration for
--- `package.searchers`.
---
--- First `require` queries `package.preload[modname]`. If it has a value,
--- this value (which should be a function) is the loader. Otherwise `require`
--- searches for a Lua loader using the path stored in `package.path`. If
--- that also fails, it searches for a C loader using the path stored in
--- `package.cpath`. If that also fails, it tries an *all-in-one* loader (see
--- `package.loaders`).
---
--- Once a loader is found, `require` calls the loader with a two argument:
--- `modname` and an extra value dependent on how it got the loader. (If the
--- loader came from a file, this extra value is the file name.) If the loader
--- returns any non-nil value, require assigns the returned value to
--- `package.loaded[modname]`. If the loader does not return a non-nil value and
--- has not assigned any value to `package.loaded[modname]`, then `require`
--- assigns true to this entry. In any case, require returns the final value of
--- `package.loaded[modname]`.
---
--- If there is any error loading or running the module, or if it cannot find
--- any loader for the module, then `require` raises an error.
---@param modname string
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-require)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-require"]
function require(modname) end
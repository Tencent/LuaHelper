---@class debug @This library provides the functionality of the debug interface to Lua programs. You should exert care when using this library. Several of its functions violate basic assumptions about Lua code (e.g., that variables local to a function cannot be accessed from outside; that userdata metatables cannot be changed by Lua code; that Lua programs do not crash) and therefore can compromise otherwise secure code. Moreover, some functions in this library may be slow. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.10)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.10"])
debug = {}

---@class debuginfo
---@field name            string
---@field namewhat        string
---@field source          string
---@field short_src       string
---@field linedefined     integer
---@field lastlinedefined integer
---@field what            string
---@field currentline     integer
---@field istailcall      boolean
---@field nups            integer
---@field nparams         integer
---@field isvararg        boolean
---@field func            function
---@field ftransfer       integer
---@field ntransfer       integer
---@field activelines     table

--- Enters an interactive mode with the user, running each string that the user
--- enters. Using simple commands and other debug facilities, the user can
--- inspect global and local variables, change their values, evaluate
--- expressions, and so on. A line containing only the word `cont` finishes this
--- function, so that the caller continues its execution.
---
--- Note that commands for `debug.debug` are not lexically nested within any
--- function, and so have no direct access to local variables.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.debug)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.debug"])
function debug.debug() end

--- Returns the environment of object o.
---@version lua5.1
---@param o any
---@return table
function debug.getfenv(o) end

--- Returns the current hook settings of the thread, as three values: the
--- current hook function, the current hook mask, and the current hook count
--- (as set by the `debug.sethook` function).
---@param co? thread
---@return function hook
---@return string mask
---@return integer count
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.gethook)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.gethook"])
function debug.gethook(co) end

---@alias infowhat string
---|'"n"'     # ---#DESTAIL 'infowhat.n'
---|'"S"'     # ---#DESTAIL 'infowhat.S'
---|'"l"'     # ---#DESTAIL 'infowhat.l'
---|'"t"'     # ---#DESTAIL 'infowhat.t'
---|'"u"'     # ---#DESTAIL 'infowhat.u'
---|'"f"'     # ---#DESTAIL 'infowhat.f'
---|'"r"'     # ---#DESTAIL 'infowhat.r'
---|'"L"'     # ---#DESTAIL 'infowhat.L'

--- Returns a table with information about a function. You can give the
--- function directly, or you can give a number as the value of `f`,
--- which means the function running at level `f` of the call stack
--- of the given thread: level 0 is the current function (`getinfo` itself);
--- level 1 is the function that called `getinfo` (except for tail calls, which
--- do not count on the stack); and so on. If `f` is a number larger than
--- the number of active functions, then `getinfo` returns **nil**.
---
--- The returned table can contain all the fields returned by `lua_getinfo`,
--- with the string `what` describing which fields to fill in. The default for
--- `what` is to get all information available, except the table of valid
--- lines. If present, the option '`f`' adds a field named `func` with the
--- function itself. If present, the option '`L`' adds a field named
--- `activelines` with the table of valid lines.
---
--- For instance, the expression `debug.getinfo(1,"n").name` returns a table
--- with a name for the current function, if a reasonable name can be found,
--- and the expression `debug.getinfo(print)` returns a table with all available
--- information about the `print` function.
---@overload fun(f: integer|function, what?: string):debuginfo
---@param thread thread
---@param f      integer|function
---@param what?  infowhat
---@return debuginfo
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getinfo)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getinfo"])
function debug.getinfo(thread, f, what) end

--- This function returns the name and the value of the local variable with
--- index `local` of the function at level `level f` of the stack. This function
--- accesses not only explicit local variables, but also parameters,
--- temporaries, etc.
---
--- The first parameter or local variable has index 1, and so on, following the
--- order that they are declared in the code, counting only the variables that
--- are active in the current scope of the function. Negative indices refer to
--- vararg parameters; -1 is the first vararg parameter. The function returns
--- **nil** if there is no variable with the given index, and raises an error
--- when called with a level out of range. (You can call `debug.getinfo` to
--- check whether the level is valid.)
---
--- Variable names starting with '(' (open parenthesis) represent variables with
--- no known names (internal variables such as loop control variables, and
--- variables from chunks saved without debug information).
---
--- The parameter `f` may also be a function. In that case, `getlocal` returns
--- only the name of function parameters.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getlocal)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getlocal"])
---@overload fun(level: integer, index: integer):string, any
---@param thread  thread
---@param level   integer
---@param index   integer
---@return string name
---@return any    value
function debug.getlocal(thread, level, index) end

--- Returns the metatable of the given `value` or **nil** if it does not have a metatable.
---@param object any
---@return table metatable
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getmetatable)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getmetatable"])
function debug.getmetatable(object) end

---Returns the registry table.
---@return table
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getregistry)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getregistry"])
function debug.getregistry() end


--- Returns the `n`-th user value associated to the userdata `u` plus a boolean, **false** if the userdata does not have that value.
---@param u userdata
---@param n number
---@return boolean
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getuservalue)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getuservalue"])
function debug.getuservalue(u, n) end

-- This function returns the name and the value of the upvalue with index up of the function f. The function returns fail if there is no upvalue with the given index.
--
--(For Lua functions, upvalues are the external local variables that the function uses, and that are consequently included in its closure.)
--    
--For C functions, this function uses the empty string "" as a name for all upvalues.
-- 
--Variable name '?' (interrogation mark) represents variables with no known names (variables from chunks saved without debug information).
---@param f  function
---@param up integer
---@return string name
---@return any    value
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.getupvalue)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.getupvalue"])
function debug.getupvalue(f, up) end

-- Sets the environment of the given `object` to the given `table`. Returns `object`.
---@version lua5.1
---@generic T
---@param object T
---@param env    table
---@return T object
function debug.setfenv(object, env) end

---@alias hookmask string
---|'"c"' # the hook is called every time Lua calls a function;
---|'"r"' # the hook is called every time Lua returns from a function;
---|'"l"' # the hook is called every time Lua enters a new line of code.

--- Sets the given function as a hook. The string `mask` and the number `count`
--- describe when the hook will be called. The string mask may have any
--- combination of the following characters, with the given meaning:
---
--- * `"c"`: the hook is called every time Lua calls a function;
--- * `"r"`: the hook is called every time Lua returns from a function;
--- * `"l"`: the hook is called every time Lua enters a new line of code.
---
--- Moreover, with a `count` different from zero, the hook is called after every
--- `count` instructions.
---
--- When called without arguments, `debug.sethook` turns off the hook.
---
--- When the hook is called, its first parameter is a string describing
--- the event that has triggered its call: `"call"`, (or `"tail
--- call"`), `"return"`, `"line"`, and `"count"`. For line events, the hook also
--- gets the new line number as its second parameter. Inside a hook, you can
--- call `getinfo` with level 2 to get more information about the running
--- function (level 0 is the `getinfo` function, and level 1 is the hook
--- function)
---@overload fun(hook: function, mask: string, count?: integer)
---@param thread thread
---@param hook   function
---@param mask   hookmask @"c" or "r" or "l"
---@param count? integer
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.sethook)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.sethook"])
function debug.sethook(thread, hook, mask, count) end

--- This function assigns the value `value` to the local variable with
--- index `local` of the function at level `level` of the stack. The function
--- returns **nil** if there is no local variable with the given index, and
--- raises an error when called with a `level` out of range. (You can call
--- `getinfo` to check whether the level is valid.) Otherwise, it returns the
--- name of the local variable.
---@overload fun(level: integer, index: integer, value: any):string
---@param thread thread
---@param level  integer
---@param index  integer
---@param value  any
---@return string name
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.setlocal)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.setlocal"])
function debug.setlocal(thread, level, index, value) end

--- Sets the metatable for the given `object` to the given `table` (which can be **nil**). Returns value.
---@generic T
---@param value T
---@param meta  table
---@return T value
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.setmetatable)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.setmetatable"])
function debug.setmetatable(value, meta) end

--- This function assigns the value `value` to the upvalue with index `up`
--- of the function `f`. The function returns **nil** if there is no upvalue
--- with the given index. Otherwise, it returns the name of the upvalue.
---@param f     function
---@param up    integer
---@param value any
---@return string name
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.setupvalue)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.setupvalue"])
function debug.setupvalue(f, up, value) end

--- Sets the given *value* as the *n*-th associated to the given *udata*. *udata* must be a full userdata.
---
--- Returns *udata*, or **nil** if the userdata does not have that value.
---@param udata userdata
---@param value any
---@param n     integer
---@return userdata udata
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.setuservalue)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.setuservalue"])
function debug.setuservalue(udata, value, n) end


--- If *message* is present but is neither a string nor **nil**, this function
--- returns `message` without further processing. Otherwise, it returns a string
--- with a traceback of the call stack. The optional *message* string is
--- appended at the beginning of the traceback. An optional level number
--- `tells` at which level to start the traceback (default is 1, the function
--- c alling `traceback`).
---@param thread   thread
---@param message? any
---@param level?   integer
---@return string  message
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.traceback)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.traceback"])
function debug.traceback(thread, message, level) end

--- Returns a unique identifier (as a light userdata) for the upvalue numbered
--- `n` from the given function.
---
--- These unique identifiers allow a program to check whether different
--- closures share upvalues. Lua closures that share an upvalue (that is, that
--- access a same external local variable) will return identical ids for those
--- upvalue indices.
---@param f fun():number
---@param n integer
---@return lightuserdata id
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.upvalueid)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.upvalueid"])
function debug.upvalueid(f, n) end

--- Make the *n1*-th upvalue of the Lua closure f1 refer to the *n2*-th upvalue of the Lua closure f2.
---@param f1 function
---@param n1 integer
---@param f2 function
---@param n2 integer
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-debug.upvaluejoin)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-debug.upvaluejoin"])
function debug.upvaluejoin(f1, n1, f2, n2) end
---@class coroutinelib @Lua supports coroutines, also called collaborative multithreading. A coroutine in Lua represents an independent thread of execution. Unlike threads in multithread systems, however, a coroutine only suspends its execution by explicitly calling a yield function. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2.6)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/2.6"])
coroutine = {}

--- Creates a new coroutine, with body *f*. *f* must be a Lua function. Returns this new coroutine, an object with type `"thread"`.
---@param f fun():thread
---@return thread
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.create)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.create"])
function coroutine.create(f) end

--- Returns true when the running coroutine can yield.
---
--- A running coroutine is yieldable if it is not the main thread and it is not inside a non-yieldable C function.
---@version lua5.4
---@param co? thread
---@return boolean
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.isyieldable)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.isyieldable"])
function coroutine.isyieldable(co) end

-- Closes coroutine *co*, that is, closes all its pending to-be-closed variables and puts the coroutine in a dead state. The given coroutine must be dead or suspended. In case of error (either the original error that stopped the coroutine or errors in closing methods), returns `false` plus the error object; otherwise returns `true`.
---@version lua5.4
---@param co thread
---@return boolean noerror
---@return any errorobject
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.close)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.close"])
function coroutine.close(co) end


--- Starts or continues the execution of coroutine `co`. The first time you resume a coroutine, it starts running its body. The values `val1`, ... are passed as the arguments to the body function. If the coroutine has yielded, `resume` restarts it; the values `val1`, ... are passed as the results from the yield.
---
--- If the coroutine runs without any errors, `resume` returns **true** plus any values passed to `yield` (when the coroutine yields) or any values returned by the body function (when the coroutine terminates). If there is any error, `resume` returns **false** plus the error message.
---@param co    thread
---@param val1? any
---@return boolean success
---@return any result
---@return ...
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.resume)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.resume"])
function coroutine.resume(co, val1, ...) end

--- Returns the running coroutine plus a boolean, true when the running coroutine is the main one.
---@return thread running
---@return boolean ismain
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.running)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.running"])
function coroutine.running() end

--- Returns the status of coroutine `co`, as a string: "`running`", if the coroutine is running (that is, it called `status`); "`suspended`", if the coroutine is suspended in a call to `yield`, or if it has not started running yet; "`normal`" if the coroutine is active but not running (that is, it has resumed another coroutine); and "`dead`" if the coroutine has finished its body function, or if it has stopped with an error.
---@param co thread
---@return string
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.status)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.status"])
function coroutine.status(co) end

--- Creates a new coroutine, with body `f`. `f` must be a Lua function. Returns
--- a function that resumes the coroutine each time it is called. Any arguments
--- passed to the function behave as the extra arguments to `resume`. Returns
--- the same values returned by `resume`, except the first
--- boolean. In case of error, propagates the error.
---@param f fun():thread
---@return fun():any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.wrap)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.wrap"])
function coroutine.wrap(f) end

--- Suspends the execution of the calling coroutine. Any arguments to `yield` are passed as extra results to `resume`.
---@return any
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-coroutine.yield)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-coroutine.yield"])
function coroutine.yield(...) end

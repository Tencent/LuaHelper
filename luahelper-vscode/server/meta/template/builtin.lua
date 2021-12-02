---@class any @any type

---@class nil:any @The type `nil` has one single value `nil`, whose main property is to be different from any other value. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@class boolean: any @The type `boolean` has two values, `false` and `true`. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@class number:any @The type `number` uses two internal representations, or two subtypes, one called *integer* and the other called *float*. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@alias integer number @integer numbers

---@class thread: any @The type *thread* represents independent threads of execution and it is used to implement coroutines. Lua threads are not related to operating-system threads. Lua supports coroutines on all systems, even thosethat do not support threads natively. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@class table: any @The type *table* implements associative arrays, that is, arrays that can have as indices not only numbers, but any Lua value except **nil** and NaN.(*Not a Number* is a special floating-point value used by the IEEE 754 standard to represent undefined or unrepresentable numerical results, such as `0/0`.) Tables can be heterogeneous; that is, they can contain values of all types (except **nil**). Any key with value **nil** is not considered part oft he table. Conversely, any key that is not part of a table has an a ssociated value **nil**. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@class string: any

---@class void

---@class userdata: any @The type `userdata` is provided to allow arbitrary C data to be stored in Lua variables. A `userdata` value represents a block of raw memory. There are two kinds of `userdata`: `full userdata`, which is an object with a block of memory managed by Lua, and `light userdata`, which is simply a C pointer value. Userdata has no predefined operations in Lua, except assignment and identity test. By using metatables, the programmer can define operations for `full userdata` values. Userdata values cannot be created or modified in Lua, only through the C API. This guarantees the integrity of data owned by the host program and C libraries. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#2)  |  [`View local doc`](command:extension.lua.doc?["en-us/54/manual.html/2"])

---@class lightuserdata: userdata

---@class function: any @Lua can call (and manipulate) functions written in Lua and functions written in C. Both are represented by the type *function*.

---@class os @This library is implemented through table os. [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.8)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.8"])
os = {}

--- Returns an approximation of the amount in seconds of CPU time used by the program.
---@return number
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.clock)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.clock"])
function os.clock() end

---@class osdate
---@field year  integer
---@field month integer
---@field day   integer
---@field hour  integer
---@field min   integer
---@field sec   integer
---@field yday  integer
---@field isdst boolean

--- Returns a string or a table containing date and time, formatted according
--- to the given string `format`.
---
--- If the `time` argument is present, this is the time to be formatted (see
--- the `os.time` function for a description of this value). Otherwise,
--- `date` formats the current time.
---
--- If `format` starts with '`!`', then the date is formatted in Coordinated
--- Universal Time. After this optional character, if `format` is the string
--- "`*t`", then `date` returns a table with the following fields:
---
--- **`year`** (four digits)
--- **`month`** (1–12)
--- **`day`** (1-31)
--- **`hour`** (0-23)
--- **`min`** (0-59)
--- **`sec`** (0-61), due to leap seconds
--- **`wday`** (weekday, 1–7, Sunday is 1)
--- **`yday`** (day of the year, 1–366)
--- **`isdst`** (daylight saving flag, a boolean). This last field may be absent
--- if the information is not available.
---
--- If `format` is not "`*t`", then `date` returns the date as a string,
--- formatted according to the same rules as the ISO C function `strftime`.
---
--- When called without arguments, `date` returns a reasonable date and time
--- representation that depends on the host system and on the current locale.
--- (More specifically, `os.date()` is equivalent to `os.date("%c")`.)
---
--- On non-POSIX systems, this function may be not thread safe because of its
--- reliance on C function `gmtime` and C function `localtime`.
---@param format? string
---@param time?   integer
---@return string|osdate
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.date)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.date"])
function os.date(format, time) end

--- Returns the difference, in seconds, from time `t1` to time `t2`. (where the
--- times are values returned by `os.time`). In POSIX, Windows, and some other
--- systems, this value is exactly `t2`-`t1`.
---@param t2 integer
---@param t1 integer
---@return integer
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.difftime)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.difftime"])
function os.difftime(t2, t1) end

--- This function is equivalent to the C function `system`. It passes `command`
--- to be executed by an operating system shell. Its first result is **true** if
--- the command terminated successfully, or **nil** otherwise. After this first
--- result the function returns a string plus a number, as follows:
---
--- **"exit"**: the command terminated normally; the following number is the
--- exit status of the command.
--- **"signal"**: the command was terminated by a signal; the following number
--- is the signal that terminated the command.
---
--- When called without a command, `os.execute` returns a boolean that is true
--- if a shell is available.
---@param command string
---@return boolean?  suc
---@return string?  @"exit" or "signal"
---@return integer?  code
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.execute)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.execute"])
function os.execute(command) end

--- Calls the ISO C function `exit` to terminate the host program. If `code` is
--- **true**, the returned status is `EXIT_SUCCESS`; if `code` is **false**, the
--- returned status is `EXIT_FAILURE`; if `code` is a number, the returned
--- status is this number. The default value for `code` is **true**.
---
--- If the optional second argument `close` is true, closes the Lua state before
--- exiting.
---@param code?  boolean|integer
---@param close? boolean
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.exit)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.exit"])
function os.exit(code, close) end

--- Returns the value of the process environment variable `varname`, or
--- **nil** if the variable is not defined.
---@param varname string
---@return string
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.getenv)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.getenv"])
function os.getenv(varname) end

--- Deletes the file (or empty directory, on POSIX systems) with the given name.
--- If this function fails, it returns **nil**, plus a string describing the
--- error and the error code. Otherwise, it returns true.
---@param filename string
---@return boolean suc
---@return string? errmsg
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.remove)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.remove"])
function os.remove(filename) end

--- Renames the file or directory named `oldname` to `newname`. If this function
--- fails, it returns **nil**, plus a string describing the error and the error
--- code. Otherwise, it returns true.
---@param oldname string
---@param newname string
---@return boolean suc
---@return string? errmsg
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.rename)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.rename"])
function os.rename(oldname, newname) end

---@alias localecategory
---| '"all"'
---| '"collate"'
---| '"ctype"'
---| '"monetary"'
---| '"numeric"'
---| '"time"'

--- Sets the current locale of the program. `locale` is a system-dependent
--- string specifying a locale; `category` is an optional string describing
--- which category to change: `"all"`, `"collate"`, `"ctype"`, `"monetary"`,
--- `"numeric"`, or `"time"`; the default category is `"all"`. The function
--- returns the name of the new locale, or **nil** if the request cannot be
--- honored.
---
--- If `locale` is the empty string, the current locale is set to an
--- implementation-defined native locale. If `locale` is the string "`C`",
--- the current locale is set to the standard C locale.
---
--- When called with **nil** as the first argument, this function only returns
--- the name of the current locale for the given category.
---
--- This function may be not thread safe because of its reliance on C function
--- `setlocale`.
---@param locale    string|nil
---@param category? localecategory @"all" or "collate" or "ctype" or "monetary" or "numeric" or "time"
---@return string localecategory
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.setlocale)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.setlocale"])
function os.setlocale(locale, category) end

--- Returns the current time when called without arguments, or a time
--- representing the date and time specified by the given table. This table
--- must have fields `year`, `month`, and `day`, and may have fields `hour`
--- (default is 12), `min` (default is 0), `sec` (default is 0), and `isdst`
--- (default is **nil**). Other fields are ignored. For a description of these
--- fields, see the `os.date` function.
---
--- When the function is called, the values in these fields do not need to be
--- inside their valid ranges. For instance, if `sec` is -10, it means 10 seconds
--- before the time specified by the other fields; if `hour` is 1000, it means
--- 1000 hours after the time specified by the other fields.
---
--- The returned value is a number, whose meaning depends on your system. In
--- POSIX, Windows, and some other systems, this number counts the number of
--- seconds since some given start time (the "epoch"). In other systems, the
--- meaning is not specified, and the number returned by `time` can be used only
--- as an argument to `os.date` and `os.difftime`.
---
--- When called with a table, `os.time` also normalizes all the fields
--- documented in the `os.date` function, so that they represent the same time
--- as before the call but with values inside their valid ranges.
---@param date? osdate
---@return integer
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.time)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.time"])
function os.time(date) end

--- Returns a string with a file name that can be used for a temporary
--- file. The file must be explicitly opened before its use and explicitly
--- removed when no longer needed.
---
--- On some systems (POSIX), this function also creates a file with that
--- name, to avoid security risks. (Someone else might create the file with
--- wrong permissions in the time between getting the name and creating the
--- file.) You still have to open the file to use it and to remove it (even
--- if you do not use it).
---
--- When possible, you may prefer to use `io.tmpfile`, which automatically
--- removes the file when the program ends.
---@return string
---[`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-os.tmpname)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-os.tmpname"])
function os.tmpname() end
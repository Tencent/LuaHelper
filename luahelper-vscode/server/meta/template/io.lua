---@class io @The I/O library provides two different styles for file manipulation. The first one uses implicit file handles; that is, there are operations to set a default input file and a default output file, and all input/output operations are done over these default files. The second style uses explicit file handles. -- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#6.8)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/6.8"])
---@field stdin  file @Standard in.
---@field stdout file @Standard out.
---@field stderr file @Standard err.
io = {}

--- Equivalent to `file:close()`. Without a file, closes the default output file.
---@param file? file
---@return boolean?  suc
---@return string? @"exit" or "signal"
---@return integer?  code
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.close)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.close"])
function io.close(file) end

--- Equivalent to `io.output():flush()`.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.flush)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.flush"])
function io.flush() end

--- When called with a file name, it opens the named file (in text mode), and
--- sets its handle as the default input file. When called with a file handle,
--- it simply sets this file handle as the default input file. When called
--- without parameters, it returns the current default input file.
---
--- In case of errors this function raises the error, instead of returning an error code.
---@param file? string|file
---@return file
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.input)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.input"])
function io.input(file) end

--- Opens the given file name in read mode and returns an iterator function
--- works like `file:lines(...)` over the opened file. When the iterator
--- function detects the end of file, it returns no values (to finish the loop)
--- and automatically closes the file.
---
--- The call `io.lines()` (with no file name) is equivalent to `io.input():lines
--- ()`; that is, it iterates over the lines of the default
--- input file. In this case, the iterator does not close the file when the loop
--- ends.
---
--- In case of errors this function raises the error, instead of returning an error code.
---@param filename? string
---@return fun():string|number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.lines)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.lines"])
function io.lines(filename, ...) end

---@alias openmode
---| '"r"'   # read mode (the default);
---| '"w"'   # write mode;
---| '"a"'   # append mode;
---| '"r+"'  # update mode, all previous data is preserved;
---| '"w+"'  # update mode, all previous data is erased;
---| '"a+"'  # append update mode, previous data is preserved, writing is only allowed at the end of file.
---| '"rb"'  # read mode(in binary mode);
---| '"wb"'  # write mode(in binary mode);
---| '"ab"'  # append mode(in binary mode);
---| '"r+b"' # update mode, all previous data is preserved(in binary mode);
---| '"w+b"' # update mode, all previous data is erased(in binary mode);
---| '"a+b"' # append update mode, previous data is preserved, writing is only allowed at the end of file(in binary mode).

--- This function opens a file, in the mode specified in the string `mode`.  In
--- case of success, it returns a new file handle. The `mode` string can be
--- any of the following:
---
--- **"r"**: read mode (the default);
--- **"w"**: write mode;
--- **"a"**: append mode;
--- **"r+"**: update mode, all previous data is preserved;
--- **"w+"**: update mode, all previous data is erased;
--- **"a+"**: append update mode, previous data is preserved, writing is only
--- allowed at the end of file.
---
--- The `mode` string can also have a '`b`' at the end, which is needed in
--- some systems to open the file in binary mode.
---@param filename string
---@param mode openmode
---@return file
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.open)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.open"])
function io.open(filename, mode) end

--- Similar to `io.input`, but operates over the default output file.
---@param file? string|file
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.output)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.output"])
function io.output(file) end


---@alias popenmode
---| '"r"' # read data from this program(default)
---| '"w"' # write data to this program

--- This function is system dependent and is not available on all platforms.
---
--- Starts program `prog` in a separated process and returns a file handle that
--- you can use to read data from this program (if `mode` is "`r`", the default)
--- or to write data to this program (if `mode` is "`w`").
---@param prog  string
---@param mode? popenmode @"r" or "w"
---@return file
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.popen)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.popen"])
function io.popen(prog, mode) end

--- Equivalent to `io.input():read(···)`.
---@return string|number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.read)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.read"])
function io.read(...) end

--- In case of success, returns a handle for a temporary file. This file is opened in update mode and it is automatically removed when the program ends.
---@return file
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.tmpfile)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.tmpfile"])
function io.tmpfile() end

---@alias filetype
---| '"file"'        # ---#DESTAIL 'filetype.file'
---| '"closed file"' # ---#DESTAIL 'filetype.closed file'
---| 'nil'           # ---#DESTAIL 'filetype.nil'

--- Checks whether `obj` is a valid file handle. Returns the string "`file`"
--- if `obj` is an open file handle, "`closed file`" if `obj` is a closed file
--- handle, or **nil** if `obj` is not a file handle.
---@param file file | string
---@return filetype @"file" or "closed file" or "nil"
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.type)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.type"])
function io.type(file) end

-- Equivalent to `io.output():write(...)`.
---@return file 
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-io.write)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-io.write"])
function io.write(...) end

---@class file @File object [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file"])
local file = {}

--- Closes `file`. Note that files are automatically closed when their
--- handles are garbage collected, but that takes an unpredictable amount of
--- time to happen.
---
--- When closing a file handle created with `io.popen`, `file:close` returns the
--- same values returned by `os.execute`.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:close)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:close"])
function file:close() end

-- Saves any written data to `file`.
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:flush)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:flush"])
function file:flush() end

--- Returns an iterator function that, each time it is called, reads the file
--- according to the given formats. When no format is given, uses "l" as a
--- default. As an example, the construction
--- `for c in file:lines(1) do *body* end`
--- will iterate over all characters of the file, starting at the current
--- position. Unlike `io.lines`, this function does not close the file when the
--- loop ends.
---
--- In case of errors this function raises the error, instead of returning an
--- error code.
---@return fun():string|number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:lines)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:lines"])
function file:lines(...) end

--- Reads the file `file`, according to the given formats, which specify
--- what to read. For each format, the function returns a string or a number
--- with the characters read, or **nil** if it cannot read data with the
--- specified format. (In this latter case, the function does not read
--- subsequent formats.) When called without parameters, it uses a default
--- format that reads the next line (see below).
----
--- The available formats are:
--- **"n"**: reads a numeral and returns it as a float or an integer, following
--- the lexical conventions of Lua. (The numeral may have leading spaces and a
--- sign.) This format always reads the longest input sequence that is a valid
--- prefix for a numeral; if that prefix does not form a valid numeral (e.g., an
--- empty string, "`0x`", or "`3.4e-`"), it is discarded and the format returns
--- **nil**;
--- **"a"**: reads the whole file, starting at the current position. On end of
--- file, it returns the empty string;
--- **"l"**: reads the next line skipping the end of line, returning **nil** on
--- end of file. This is the default format.
--- **"L"**: reads the next line keeping the end-of-line character (if present),
--- returning **nil** on end of file;
--- *number*: reads a string with up to this number of bytes, returning **nil**
--- on end of file. If `number` is zero, it reads nothing and returns an
--- empty string, or **nil** on end of file.
---@return string|number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:read)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:read"])
function file:read(...) end

---@alias seekwhence
---| '"set"' # base is position 0 (beginning of the file);
---| '"cur"' # base is current position;
---| '"end"' # base is end of file;

--- Sets and gets the file position, measured from the beginning of the
--- file, to the position given by `offset` plus a base specified by the string
--- `whence`, as follows:
--- **"set"**: base is position 0 (beginning of the file);
--- **"cur"**: base is current position;
--- **"end"**: base is end of file;
---
--- In case of success, `seek` returns the final file position, measured in
--- bytes from the beginning of the file. If `seek` fails, it returns **nil**,
--- plus a string describing the error.
---
--- The default value for `whence` is "`cur`", and for `offset` is 0. Therefore,
--- the call `file:seek()` returns the current file position, without changing
--- it; the call `file:seek("set")` sets the position to the beginning of the
--- file (and returns 0); and the call `file:seek("end")` sets the position
--- to the end of the file, and returns its size.
---@param whence? seekwhence @ "set" or "cur" or "end"
---@param offset? number
---@return integer @offset
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:seek)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:seek"])
function file:seek(whence, offset) end

---@alias vbuf
---| '"no"'   # no buffering; the result of any output operation appears immediately.
---| '"full"' # full buffering;
---| '"line"' # line buffering;

--- Sets the buffering mode for an output file. There are three available
--- modes:
--- **"no"**: no buffering; the result of any output operation appears
--- immediately.
--- **"full"**: full buffering; output operation is performed only when the
--- buffer is full (or when you explicitly `flush` the file (see `io.flush`)).
--- **"line"**: line buffering; output is buffered until a newline is output or
--- there is any input from some special files (such as a terminal device).
---
--- For the last two cases, `size` specifies the size of the buffer, in
--- bytes. The default is an appropriate size.
---@param mode? vbuf @ "no" or "full" or "line"
---@param size? number
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:setvbuf)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:setvbuf"])
function file:setvbuf(mode, size) end

--- Writes the value of each of its arguments to the `file`. The arguments
--- must be strings or numbers.
---
--- In case of success, this function returns `file`. Otherwise it returns
--- **nil** plus a string describing the error.
---@return file
---@return string @errmsg
-- [`View online doc`](https://www.lua.org/manual/5.4/manual.html#pdf-file:write)  |  [`View local doc`](command:extension.luahelper.doc?["en-us/54/manual.html/pdf-file:write"])
function file:write(...) end


package common

// 当为emacs、vim等插件时候系统函数的提示，这些插件不不好系统库的lua文件

// FuncParamInfo 函数的参数信息
type FuncParamInfo struct {
	Label         string
	Documentation string
}

// SystemNoticeInfo 系统自带的库，提示信息设置
type SystemNoticeInfo struct {
	Detail        string // 提示展示信息
	Documentation string // 进一步信息
	FuncParamVec  []FuncParamInfo
}

// SystemModuleVar 模块内的变量信息
type SystemModuleVar struct {
	Label         string
	Detail        string // 提示展示信息
	Documentation string // 进一步信息
}

// OneModuleInfo 一个模块所有信息
type OneModuleInfo struct {
	Detail        string // 提示展示信息
	Documentation string // 进一步信息
	ModuleFuncMap map[string]*SystemNoticeInfo
	ModuleVarVec  map[string]*SystemModuleVar // 模块内所有变量
}

func createSysVarInfo(strName string, noticeInfo *SystemNoticeInfo, moduleInfo *OneModuleInfo) *VarInfo {
	locVar := &VarInfo{
		ReferExp:  nil,
		ReferInfo: nil,
		ReferFunc: nil,
		SubMaps:   nil,
		VarType:   LuaTypeFunc,
		IsParam:   false,
		IsUse:   false,
	}

	extraGlobal := &ExtraGlobal{
		Prev:      nil,
		StrProPre: "", // 表示是否为协议的前缀 为c2s 或是s2s
		FuncLv:    0,  // 函数的层级，主函数层级为0，子函数+1
		ScopeLv:   0,  // 所在的scop层数，相对于自己所处的func而言
		GFlag:     true,
	}

	locVar.ExtraGlobal = extraGlobal

	extraSystem := &ExtraSystem{
		SysNoticeInfo: noticeInfo,
		SysModuleInfo: moduleInfo,
	}
	locVar.ExtraGlobal.ExtraSystem = extraSystem
	return locVar
}

// 创新系统的变量，非函数
func createSysModuleVarInfo(strName string, sysModuleVar *SystemModuleVar) *VarInfo {
	varInfo := createSysVarInfo(strName, nil, nil)
	varInfo.ExtraGlobal.ExtraSystem.SysModuleVar = sysModuleVar
	return varInfo
}

// 创建一个系统的VarInfo
func (g *GlobalConfig) insertSysVarInfo(strName string, noticeInfo *SystemNoticeInfo, moduleInfo *OneModuleInfo) {
	g.SysVarMap[strName] = createSysVarInfo(strName, noticeInfo, moduleInfo)

	if moduleInfo != nil {
		g.insertOneModuleSubVars(strName, moduleInfo)
	}
}

// 把模块的所有成员，放入进来
func (g *GlobalConfig) insertOneModuleSubVars(moduleName string, oneModule *OneModuleInfo) {
	locVar, ok := g.SysVarMap[moduleName]
	if !ok {
		return
	}

	locVar.SubMaps = map[string]*VarInfo{}
	for strName, subOne := range oneModule.ModuleFuncMap {
		subVar := createSysVarInfo(strName, subOne, nil)
		locVar.SubMaps[strName] = subVar
	}

	for strName, subModuleOne := range oneModule.ModuleVarVec {
		subVar := createSysModuleVarInfo(strName, subModuleOne)
		locVar.SubMaps[strName] = subVar
	}
}

// InitSystemTips 初始化系统的函数的提示
func (g *GlobalConfig) InitSystemTips() {
	g.SysVarMap = map[string]*VarInfo{}

	g.SystemTipsMap = map[string]SystemNoticeInfo{}
	assertNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] assert(v, message)",
		Documentation: "Calls error if the value of its argument `v` is false (i.e., **nil** or **false**); otherwise, returns all its arguments. In case of error, `message` is the error object; when absent, it defaults to \"assertion failed!\"",
		FuncParamVec:  []FuncParamInfo{{"v", "v : any"}, {"message", "message : string"}},
	}
	g.SystemTipsMap["assert"] = assertNoticeInfo
	g.insertSysVarInfo("assert", &assertNoticeInfo, nil)

	collectgarbageNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] collectgarbage(opt, arg)",
		Documentation: "This function is a generic interface to the garbage collector. It performs different functions according to its first argument, `opt`",
		FuncParamVec:  []FuncParamInfo{{"opt", "opt : string"}, {"arg", "arg : string"}},
	}
	g.SystemTipsMap["collectgarbage"] = collectgarbageNoticeInfo
	g.insertSysVarInfo("collectgarbage", &collectgarbageNoticeInfo, nil)

	dofileNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] dofile(filename)",
		Documentation: "Opens the named file and executes its contents as a Lua chunk",
		FuncParamVec:  []FuncParamInfo{{"filename", "filename : string"}},
	}
	g.SystemTipsMap["dofile"] = dofileNoticeInfo
	g.insertSysVarInfo("dofile", &dofileNoticeInfo, nil)

	errorNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] error(message, level)",
		Documentation: "Terminates the last protected function called and returns `message` as the error object. Function `error` never returns",
		FuncParamVec:  []FuncParamInfo{{"message", "message : string"}, {"level", "level : number"}},
	}
	g.SystemTipsMap["error"] = errorNoticeInfo
	g.insertSysVarInfo("error", &errorNoticeInfo, nil)

	getmetatableNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] getmetatable(object)",
		Documentation: "If `object` does not have a metatable, returns **nil**. Otherwise, if the object's metatable has a `\"__metatable\"` field, returns the associated value. Otherwise, returns the metatable of the given object",
		FuncParamVec:  []FuncParamInfo{{"object", "object : any"}},
	}
	g.SystemTipsMap["getmetatable"] = getmetatableNoticeInfo
	g.insertSysVarInfo("getmetatable", &getmetatableNoticeInfo, nil)

	ipairsNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] ipairs(t)",
		Documentation: "Returns three values (an iterator function, the table `t`, and 0) so that the construction",
		FuncParamVec:  []FuncParamInfo{{"t", "t: table"}},
	}
	g.SystemTipsMap["ipairs"] = ipairsNoticeInfo
	g.insertSysVarInfo("ipairs", &ipairsNoticeInfo, nil)

	//g.SystemTipsMap["load(chunk, chunkname, mode, env)"] = "[_G]"
	loadNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] load(chunk, chunkname, mode, env)",
		Documentation: "Loads a chunk. If `chunk` is a string, the chunk is this string. If `chunk` is a function,`load` calls it repeatedly to get the chunk pieces. Each call to `chunk` must return a string that concatenates with previous results. A return of an empty string, **nil**, or no value signals the end of the chunk.",
		FuncParamVec: []FuncParamInfo{{"chunk", "chunk: function"}, {"chunkname", "chunkname : string"},
			{"mode", "mode : string"}, {"env", "env ：any"}},
	}
	g.SystemTipsMap["load"] = loadNoticeInfo
	g.insertSysVarInfo("load", &loadNoticeInfo, nil)

	loadfileNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] loadfile(filename, mode, env)",
		Documentation: "Similar to `load`, but gets the chunk from file `filename` or from the standard input, if no file name is given",
		FuncParamVec: []FuncParamInfo{{"filename", "filename ：string"}, {"mode", "mode ：string"},
			{"env", "env : any"}},
	}
	g.SystemTipsMap["loadfile"] = loadfileNoticeInfo
	g.insertSysVarInfo("loadfile", &loadfileNoticeInfo, nil)

	nextNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] next(table, index)",
		Documentation: "Allows a program to traverse all fields of a table. Its first argument is a table and its second argument is an index in this table. `next` returns the next index of the table and its associated value.",
		FuncParamVec:  []FuncParamInfo{{"table", "table : table"}, {"index", "index : number"}},
	}
	g.SystemTipsMap["next"] = nextNoticeInfo
	g.insertSysVarInfo("next", &nextNoticeInfo, nil)

	pairsNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] pairs(t)",
		Documentation: "for k,v in pairs(t) do *body* end",
		FuncParamVec:  []FuncParamInfo{{"t", "t: table"}},
	}
	g.SystemTipsMap["pairs"] = pairsNoticeInfo
	g.insertSysVarInfo("pairs", &pairsNoticeInfo, nil)

	pcallNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] pcall(f, arg1, ...)",
		Documentation: "Calls function `f` with the given arguments in *protected mode*",
		FuncParamVec:  []FuncParamInfo{{"f", "f: function"}, {"arg1", "arg1 : table"}, {"...)", ""}},
	}
	g.SystemTipsMap["pcall"] = pcallNoticeInfo
	g.insertSysVarInfo("pcall", &pcallNoticeInfo, nil)

	printNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] print(...)",
		Documentation: "Receives any number of arguments, and prints their values to `stdout`, using the `tostring` function to convert them to strings",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	g.SystemTipsMap["print"] = printNoticeInfo
	g.insertSysVarInfo("print", &printNoticeInfo, nil)

	rawequalNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] rawequal((v1, v2)",
		Documentation: "Checks whether `v1` is equal to `v2`, without the `__eq` metamethod. Returns a bool.",
		FuncParamVec:  []FuncParamInfo{{"v1", "v1 : any"}, {"v2", "v2 : any"}},
	}
	g.SystemTipsMap["rawequal"] = rawequalNoticeInfo
	g.insertSysVarInfo("rawequal", &rawequalNoticeInfo, nil)

	rawgetNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] rawget(table, index)",
		Documentation: "Gets the real value of `table[index]`, the `__index` metamethod. `table` must be a table; `index` may be any value.",
		FuncParamVec:  []FuncParamInfo{{"table", "table : table"}, {"index", "index : number"}},
	}
	g.SystemTipsMap["rawget"] = rawgetNoticeInfo
	g.insertSysVarInfo("rawget", &rawgetNoticeInfo, nil)

	rawlenNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] rawlen(v)",
		Documentation: "Returns the length of the object `v`, which must be a table or a string, without invoking any metamethod. Returns an integer number",
		FuncParamVec:  []FuncParamInfo{{"v", "v: string|table"}},
	}
	g.SystemTipsMap["rawlen"] = rawlenNoticeInfo
	g.insertSysVarInfo("rawlen", &rawlenNoticeInfo, nil)

	rawsetNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] rawset(table, index, value)",
		Documentation: "Sets the real value of `table[index]` to `value`",
		FuncParamVec: []FuncParamInfo{{"table", "table : table"}, {"index", "index : any"},
			{"value", "value : any"}},
	}
	g.SystemTipsMap["rawset"] = rawsetNoticeInfo
	g.insertSysVarInfo("rawset", &rawsetNoticeInfo, nil)

	requireNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] require(modname)",
		Documentation: "Loads the given module",
		FuncParamVec:  []FuncParamInfo{{"modname", "modname : string"}},
	}
	g.SystemTipsMap["require"] = requireNoticeInfo
	g.insertSysVarInfo("require", &requireNoticeInfo, nil)

	selectNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] select(index, ...)",
		Documentation: "If `index` is a number, returns all arguments after argument number, `index`. a negative number indexes from the end (-1 is the last argument).",
		FuncParamVec:  []FuncParamInfo{{"index", "index : number|string"}, {"...", ""}},
	}
	g.SystemTipsMap["select"] = selectNoticeInfo
	g.insertSysVarInfo("select", &selectNoticeInfo, nil)

	setmetatableNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] setmetatable(table, metatable)",
		Documentation: "Sets the metatable for the given table. (To change the metatable of other types from Lua code, you must use the debug library.)",
		FuncParamVec:  []FuncParamInfo{{"table", "table : table"}, {"metatable", "metatable : table"}},
	}
	g.SystemTipsMap["setmetatable"] = setmetatableNoticeInfo
	g.insertSysVarInfo("setmetatable", &setmetatableNoticeInfo, nil)

	tonumberNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] tonumber(e, base)",
		Documentation: "When called with no `base`, `tonumber` tries to convert its argument to a number. If the argument is already a number or a string convertible to a number, then `tonumber` returns this number; otherwise, it returns **nil**",
		FuncParamVec:  []FuncParamInfo{{"e", "e : any"}, {"base", "base : number"}},
	}
	g.SystemTipsMap["tonumber"] = tonumberNoticeInfo
	g.insertSysVarInfo("tonumber", &tonumberNoticeInfo, nil)

	tostringNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] tostring(v) ",
		Documentation: "Receives a value of any type and converts it to a string in a human-readable format. (For complete control of how numbers are converted, use `string.format`)",
		FuncParamVec:  []FuncParamInfo{{"v", "v : any"}},
	}
	g.SystemTipsMap["tostring"] = tostringNoticeInfo
	g.insertSysVarInfo("tostring", &tostringNoticeInfo, nil)

	typeNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] type(v)",
		Documentation: "Returns the type of its only argument, coded as a string. The possible results of this function are \"`nil`\" (a string, not the value **nil**)",
		FuncParamVec:  []FuncParamInfo{{"v", "v : any"}},
	}
	g.SystemTipsMap["type"] = typeNoticeInfo
	g.insertSysVarInfo("type", &typeNoticeInfo, nil)

	xpcallNoticeInfo := SystemNoticeInfo{
		Detail:        "[_G] xpcall(f, msgh, arg1, ...)",
		Documentation: "This function is similar to `pcall`, except that it sets a new message handler `msgh`",
		FuncParamVec:  []FuncParamInfo{{"f", "f : function"}, {"msgh", "msgh : function"}, {"arg1", "arg1 : any"}, {"...", ""}},
	}
	g.SystemTipsMap["xpcall"] = xpcallNoticeInfo
	g.insertSysVarInfo("xpcall", &xpcallNoticeInfo, nil)

	g.SystemModuleTipsMap = map[string]OneModuleInfo{}

	// 1) coroutine 模块
	coroutineModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "coroutine module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}

	coroutineModule.ModuleFuncMap["create"] = &SystemNoticeInfo{
		Detail:        "create(f)",
		Documentation: "Creates a new coroutine, with body `f`. `f` must be a Lua function. Returns this new coroutine, an object with type `\"thread\"`.",
		FuncParamVec:  []FuncParamInfo{{"f", "f : function"}},
	}
	coroutineModule.ModuleFuncMap["isyieldable"] = &SystemNoticeInfo{
		Detail:        "isyieldable()",
		Documentation: "Returns true when the running coroutine can yield",
		FuncParamVec:  []FuncParamInfo{},
	}
	coroutineModule.ModuleFuncMap["resume"] = &SystemNoticeInfo{
		Detail:        "resume(co, val1, ...)",
		Documentation: "Starts or continues the execution of coroutine `co`. The first time you resume a coroutine, it starts running its body. The values `val1`, ... are passed as the arguments to the body function. If the coroutine has yielded, `resume` restarts it; the values `val1`, ... are passed as the results from the yield.",
		FuncParamVec:  []FuncParamInfo{{"co", "co : thread"}, {"val1", "val1 : string"}, {"...", ""}},
	}

	coroutineModule.ModuleFuncMap["running"] = &SystemNoticeInfo{
		Detail:        "running()",
		Documentation: "Returns the running coroutine plus a bool, true when the running coroutine is the main one.",
		FuncParamVec:  []FuncParamInfo{},
	}
	coroutineModule.ModuleFuncMap["status"] = &SystemNoticeInfo{
		Detail:        "status(co)",
		Documentation: "Returns the status of coroutine `co`, as a string",
		FuncParamVec:  []FuncParamInfo{{"co", "co : thread"}},
	}
	coroutineModule.ModuleFuncMap["wrap"] = &SystemNoticeInfo{
		Detail:        "wrap(f)",
		Documentation: "Creates a new coroutine, with body `f`. `f` must be a Lua function. Returns a function that resumes the coroutine each time it is called.",
		FuncParamVec:  []FuncParamInfo{{"f", "f : function"}},
	}
	coroutineModule.ModuleFuncMap["yield"] = &SystemNoticeInfo{
		Detail:        "yield(...)",
		Documentation: "Suspends the execution of the calling coroutine. Any arguments to `yield` are passed as extra results to `resume`",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}

	g.SystemModuleTipsMap["coroutine"] = coroutineModule
	g.insertSysVarInfo("coroutine", nil, &coroutineModule)

	// 2) debug模块
	debugModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "debug module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}

	debugModule.ModuleFuncMap["debug"] = &SystemNoticeInfo{
		Detail:        "debug()",
		Documentation: "Enters an interactive mode with the user, running each string that the user enters",
		FuncParamVec:  []FuncParamInfo{},
	}

	debugModule.ModuleFuncMap["gethook"] = &SystemNoticeInfo{
		Detail:        "gethook(thread)",
		Documentation: "Returns the current hook settings of the thread, as three values: the current hook function, the current hook mask, and the current hook count (as set by the `debug.sethook` function)",
		FuncParamVec:  []FuncParamInfo{{"thread", "thread : thread"}},
	}
	debugModule.ModuleFuncMap["getinfo"] = &SystemNoticeInfo{
		Detail:        "getinfo(thread, f, what)",
		Documentation: "Returns a table with information about a function",
		FuncParamVec: []FuncParamInfo{{"thread", "thread : thread"}, {"f", "f : function"},
			{"what", "what : string"}},
	}
	debugModule.ModuleFuncMap["getlocal"] = &SystemNoticeInfo{
		Detail:        "getlocal(thread, f, what)",
		Documentation: "This function returns the name and the value of the local variable with index `local` of the function at level `level f` of the stack",
		FuncParamVec: []FuncParamInfo{{"thread", "thread : thread"}, {"f", "f : function"},
			{"what", "what : string"}},
	}
	debugModule.ModuleFuncMap["getmetatable"] = &SystemNoticeInfo{
		Detail:        "getmetatable(value)",
		Documentation: "Returns the metatable of the given `value` or **nil** if it does not have a metatable.",
		FuncParamVec:  []FuncParamInfo{{"value", "value : table"}},
	}
	debugModule.ModuleFuncMap["getregistry"] = &SystemNoticeInfo{
		Detail:        "getregistry()",
		Documentation: "Returns the registry table",
		FuncParamVec:  []FuncParamInfo{},
	}
	debugModule.ModuleFuncMap["getupvalue"] = &SystemNoticeInfo{
		Detail:        "getupvalue(f, up)",
		Documentation: "This function returns the name and the value of the upvalue with index `up` of the function `f`. The function returns **nil** if there is no upvalue with the given index",
		FuncParamVec:  []FuncParamInfo{{"f", "f : number"}, {"up", "up : number"}},
	}

	debugModule.ModuleFuncMap["getuservalue"] = &SystemNoticeInfo{
		Detail:        "getuservalue(u, n)",
		Documentation: "Returns the `n`-th user value associated to the userdata `u` plus a bool, **false** if the userdata does not have that value.",
		FuncParamVec:  []FuncParamInfo{{"u", "u : userdata"}, {"n", "n : number"}},
	}
	debugModule.ModuleFuncMap["sethook"] = &SystemNoticeInfo{
		Detail:        "sethook(thread, hook, mask, count)",
		Documentation: "Sets the given function as a hook.",
		FuncParamVec: []FuncParamInfo{{"thread", "thread : thread"}, {"hook", "hook : function"},
			{"mask", "mask : string"}, {"count", "count : number"}},
	}
	debugModule.ModuleFuncMap["setlocal"] = &SystemNoticeInfo{
		Detail:        "setlocal(thread, level, var, value)",
		Documentation: "This function assigns the value `value` to the local variable with index `local` of the function at level `level` of the stack. The function returns **nil** if there is no local variable with the given index, and raises an error when called with a `level` out of range. ",
		FuncParamVec: []FuncParamInfo{{"thread", "thread : thread"}, {"level", "level : number"},
			{"var", "var : string"}, {"value", "value : any"}},
	}
	debugModule.ModuleFuncMap["setmetatable"] = &SystemNoticeInfo{
		Detail:        "setmetatable(value, table)",
		Documentation: "Sets the metatable for the given `object` to the given `table` (which can be **nil**). Returns value.",
		FuncParamVec:  []FuncParamInfo{{"value", "value : any"}, {"table", "table : table"}},
	}
	debugModule.ModuleFuncMap["setupvalue"] = &SystemNoticeInfo{
		Detail:        "setupvalue(f, up, value)",
		Documentation: "Sets the metatable for the given `object` to the given `table` (which can be **nil**). Returns value.",
		FuncParamVec: []FuncParamInfo{{"f", "f : function"}, {"up", "up : number"},
			{"value", "value : any"}},
	}
	debugModule.ModuleFuncMap["setuservalue"] = &SystemNoticeInfo{
		Detail:        "setuservalue(udata, value, n)",
		Documentation: "Sets the given `value` as the `n`-th associated to the given `udata`. `udata` must be a full userdata",
		FuncParamVec: []FuncParamInfo{{"udata", "udata : userdata"}, {"value", "value : any"},
			{"n", "n : number"}},
	}
	debugModule.ModuleFuncMap["traceback"] = &SystemNoticeInfo{
		Detail:        "traceback(thread, message, level)",
		Documentation: "If `message` is present but is neither a string nor **nil**, this function returns `message` without further processing. Otherwise, it returns a string with a traceback of the call stack. ",
		FuncParamVec: []FuncParamInfo{{"thread", "thread : thread"},
			{"message", "message : string"}, {"level", "level : number"}},
	}
	debugModule.ModuleFuncMap["upvalueid"] = &SystemNoticeInfo{
		Detail:        "upvalueid(f, n)",
		Documentation: "Returns a unique identifier (as a light userdata) for the upvalue numbered `n` from the given function",
		FuncParamVec:  []FuncParamInfo{{"f", "f : function"}, {"n", "n : number"}},
	}
	debugModule.ModuleFuncMap["upvaluejoin"] = &SystemNoticeInfo{
		Detail:        "upvaluejoin(f1, n1, f2, n2)",
		Documentation: "Make the `n1`-th upvalue of the Lua closure f1 refer to the `n2`-th upvalue of the Lua closure f2",
		FuncParamVec: []FuncParamInfo{{"f1", "f1 : function"}, {"n1", "n1 : number"},
			{"f2", "f2 : function"}, {"n2", "n2 : number"}},
	}
	g.SystemModuleTipsMap["debug"] = debugModule
	g.insertSysVarInfo("debug", nil, &debugModule)

	// 3) io 模块
	ioModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "io module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	ioModule.ModuleFuncMap["close"] = &SystemNoticeInfo{
		Detail:        "close(file)",
		Documentation: "Equivalent to `file:close()`. Without a file, closes the default output file",
		FuncParamVec:  []FuncParamInfo{{"file", "file: file"}},
	}
	ioModule.ModuleFuncMap["flush"] = &SystemNoticeInfo{
		Detail:        "flush()",
		Documentation: "Equivalent to `io.output():flush()`.",
		FuncParamVec:  []FuncParamInfo{},
	}
	ioModule.ModuleFuncMap["input"] = &SystemNoticeInfo{
		Detail:        "input(file)",
		Documentation: "When called with a file name, it opens the named file (in text mode), and sets its handle as the default input file. When called with a file handle, it simply sets this file handle as the default input file. When called without parameters, it returns the current default input file",
		FuncParamVec:  []FuncParamInfo{{"file", "file: file | string"}},
	}

	ioModule.ModuleFuncMap["lines"] = &SystemNoticeInfo{
		Detail:        "lines(filename, ...)",
		Documentation: "Opens the given file name in read mode and returns an iterator function works like `file:lines(···)` over the opened file. When the iterator function detects the end of file, it returns no values (to finish the loop) and automatically closes the file.",
		FuncParamVec:  []FuncParamInfo{{"filename", "filename : string"}, {"...", ""}},
	}

	ioModule.ModuleFuncMap["open"] = &SystemNoticeInfo{
		Detail:        "open(filename, mode)",
		Documentation: "This function opens a file, in the mode specified in the string `mode`.  In case of success, it returns a new file handle",
		FuncParamVec:  []FuncParamInfo{{"filename", "filename : string"}, {"mode", "mode : string **\"r\"**: read mode (the default); **\"w\"**: write mode; **\"a\"**: append mode; **\"r+\"**: update mode, all previous data is preserved; **\"w+\"**: update mode, all previous data is erased; **\"a+\"**: append update mode, previous data is preserved, writing is only allowed at the end of file."}},
	}

	ioModule.ModuleFuncMap["output"] = &SystemNoticeInfo{
		Detail:        "output(file)",
		Documentation: "Similar to `io.input`, but operates over the default output file.",
		FuncParamVec:  []FuncParamInfo{{"file", "file : file | string"}},
	}
	ioModule.ModuleFuncMap["popen"] = &SystemNoticeInfo{
		Detail:        "io.popen(prog, mode)",
		Documentation: "Starts program `prog` in a separated process and returns a file handle that you can use to read data from this program (if `mode` is \"`r`\", the default) or to write data to this program (if `mode` is \"`w`\")",
		FuncParamVec: []FuncParamInfo{{"prog", "string "},
			{"mode", "mode : string | '\"r\"' | '\"w\"'"}},
	}
	ioModule.ModuleFuncMap["read"] = &SystemNoticeInfo{
		Detail:        "read(...)",
		Documentation: "Equivalent to `io.input():read(···)`.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	ioModule.ModuleFuncMap["tmpfile"] = &SystemNoticeInfo{
		Detail:        "tmpfile(...)",
		Documentation: "In case of success, returns a handle for a temporary file. This file is opened in update mode and it is automatically removed when the program ends.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	ioModule.ModuleFuncMap["type"] = &SystemNoticeInfo{
		Detail:        "type(obj)",
		Documentation: "Checks whether `obj` is a valid file handle. Returns the string \"`file`\".",
		FuncParamVec:  []FuncParamInfo{{"obj", "obj : string|file"}},
	}
	ioModule.ModuleFuncMap["write"] = &SystemNoticeInfo{
		Detail:        "write(...)",
		Documentation: "Equivalent to `io.output():write(···)`.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	g.SystemModuleTipsMap["io"] = ioModule
	g.insertSysVarInfo("io", nil, &ioModule)

	// 3) file 模块
	fileModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "file module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	fileModule.ModuleFuncMap["close"] = &SystemNoticeInfo{
		Detail:        "close()",
		Documentation: "Closes `file`. Note that files are automatically closed when their handles are garbage collected, but that takes an unpredictable amount of time to happen",
		FuncParamVec:  []FuncParamInfo{},
	}
	fileModule.ModuleFuncMap["flush"] = &SystemNoticeInfo{
		Detail:        "flush()",
		Documentation: "Saves any written data to `file`.",
		FuncParamVec:  []FuncParamInfo{},
	}
	fileModule.ModuleFuncMap["lines"] = &SystemNoticeInfo{
		Detail:        "lines(...)",
		Documentation: "Returns an iterator function that, each time it is called, reads the file according to the given formats. When no format is given, uses \"l\" as a default.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	fileModule.ModuleFuncMap["read"] = &SystemNoticeInfo{
		Detail:        "read(...)",
		Documentation: " Reads the file `file`, according to the given formats, which specify what to read. For each format, the function returns a string or a number with the characters read, or **nil** if it cannot read data with the specified format.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	fileModule.ModuleFuncMap["seek"] = &SystemNoticeInfo{
		Detail:        "seek(whence, offset)",
		Documentation: "Sets and gets the file position, measured from the beginning of the file, to the position given by `offset` plus a base specified by the string `whence`, as follows: **\"set\"**: base is position 0 (beginning of the file); **\"cur\"**: base is current position; **\"end\"**: base is end of file;",
		FuncParamVec: []FuncParamInfo{{"whence", "whence : string | '\"set\"' | '\"cur\"' | '\"end\"'"},
			{"offset", "offset : number"}},
	}
	fileModule.ModuleFuncMap["setvbuf"] = &SystemNoticeInfo{
		Detail:        "setvbuf(mode, size)",
		Documentation: "Sets the buffering mode for an output file.",
		FuncParamVec: []FuncParamInfo{{"mode", "mode : string | '\"no\"' | '\"full\"' | '\"line\"'"},
			{"size", "size : number"}},
	}
	fileModule.ModuleFuncMap["write"] = &SystemNoticeInfo{
		Detail:        "write(...)",
		Documentation: "Writes the value of each of its arguments to the `file`. The arguments must be strings or numbers",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}

	fileModule.ModuleVarVec = map[string]*SystemModuleVar{}
	fileModule.ModuleVarVec["stderr"] = &SystemModuleVar{"stderr", "`io.stderr`: Standard error.", ""}
	fileModule.ModuleVarVec["stdin"] = &SystemModuleVar{"stdin", "`io.stdin`: Standard in.", ""}
	fileModule.ModuleVarVec["stdout"] = &SystemModuleVar{"stdout", "`io.stdout`: Standard out.", ""}
	g.SystemModuleTipsMap["file"] = fileModule
	g.insertSysVarInfo("file", nil, &fileModule)

	// 4) math 模块
	mathModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "math module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	mathModule.ModuleFuncMap["abs"] = &SystemNoticeInfo{
		Detail:        "abs(x)",
		Documentation: "Returns the absolute value of `x`. (integer/float)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["acos"] = &SystemNoticeInfo{
		Detail:        "acos(x)",
		Documentation: "Returns the arc cosine of `x` (in radians)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["asin"] = &SystemNoticeInfo{
		Detail:        "asin(x)",
		Documentation: "Returns the arc sine of `x` (in radians)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["atan"] = &SystemNoticeInfo{
		Detail:        "atan(y, x)",
		Documentation: "Returns the arc tangent of `y/x` (in radians), but uses the signs of both parameters to find the quadrant of the result. (It also handles correctly the case of `x` being zero.)",
		FuncParamVec:  []FuncParamInfo{{"y", "y : number"}, {"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["ceil"] = &SystemNoticeInfo{
		Detail:        "ceil(x)",
		Documentation: "Returns the smallest integer larger than or equal to `x`",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["cos"] = &SystemNoticeInfo{
		Detail:        "cos(x)",
		Documentation: "Returns the cosine of `x` (assumed to be in radians)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["deg"] = &SystemNoticeInfo{
		Detail:        "deg(x)",
		Documentation: "Converts the angle `x` from radians to degrees",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["exp"] = &SystemNoticeInfo{
		Detail:        "exp(x)",
		Documentation: "Returns the value *e^x* (where e is the base of natural logarithms)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["floor"] = &SystemNoticeInfo{
		Detail:        "floor(x)",
		Documentation: "Returns the largest integer smaller than or equal to `x`",
		FuncParamVec:  []FuncParamInfo{{"x", "x :number"}},
	}

	mathModule.ModuleFuncMap["fmod"] = &SystemNoticeInfo{
		Detail:        "fmod(x, y)",
		Documentation: "Returns the remainder of the division of `x` by `y` that rounds the quotient towards zero. (integer/float)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}, {"y", "y : number"}},
	}

	mathModule.ModuleFuncMap["log"] = &SystemNoticeInfo{
		Detail:        "log(x, base)",
		Documentation: "Returns the logarithm of `x` in the given base. The default for `base` is *e* (so that the function returns the natural logarithm of `x`)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}, {"base", "base : number"}},
	}

	mathModule.ModuleFuncMap["max"] = &SystemNoticeInfo{
		Detail:        "max(x, ...)",
		Documentation: "Returns the argument with the maximum value, according to the Lua operator `<`. (integer/float))",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}, {"...", ""}},
	}

	mathModule.ModuleFuncMap["min"] = &SystemNoticeInfo{
		Detail:        "min(x, ...)",
		Documentation: "Returns the argument with the minimum value, according to the Lua operator `<`. (integer/float))",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}, {"...", ""}},
	}

	mathModule.ModuleFuncMap["modf"] = &SystemNoticeInfo{
		Detail:        "modf(x)",
		Documentation: "Returns the integral part of `x` and the fractional part of `x`. Its second result is always a float.",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["rad"] = &SystemNoticeInfo{
		Detail:        "rad(x)",
		Documentation: "Converts the angle `x` from degrees to radians",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["random"] = &SystemNoticeInfo{
		Detail:        "random(m, n)",
		Documentation: "When called without arguments, returns a pseudo-random float with uniform distribution in the range *[0,1)*. When called with two integers `m` and `n`, `math.random` returns a pseudo-random integer with uniform distribution in the range *[m, n]*. The call `math.random(n)` is equivalent to `math.random`(1,n).",
		FuncParamVec:  []FuncParamInfo{{"m", "m : number"}, {"n", "n : number"}},
	}

	mathModule.ModuleFuncMap["randomseed"] = &SystemNoticeInfo{
		Detail:        "randomseed(x)",
		Documentation: "Sets `x` as the \"seed\" for the pseudo-random generator: equal seeds produce equal sequences of numbers",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["sin"] = &SystemNoticeInfo{
		Detail:        "sin(x)",
		Documentation: "Returns the sine of `x` (assumed to be in radians)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}
	mathModule.ModuleFuncMap["sqrt"] = &SystemNoticeInfo{
		Detail:        "sqrt(x)",
		Documentation: "Returns the square root of `x`. (You can also use the expression `x^0.5` to compute this value.)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["tan"] = &SystemNoticeInfo{
		Detail:        "tan(x)",
		Documentation: "Returns the tangent of `x` (assumed to be in radians)",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["tointeger"] = &SystemNoticeInfo{
		Detail:        "tointeger(x)",
		Documentation: " If the value `x` is convertible to an integer, returns that integer. Otherwise, returns `nil`",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleFuncMap["type"] = &SystemNoticeInfo{
		Detail:        "type(x)",
		Documentation: "Returns a bool, true if and only if integer `m` is below integer `n` when they are compared as unsigned integers",
		FuncParamVec:  []FuncParamInfo{{"x", "x : number"}},
	}

	mathModule.ModuleVarVec = map[string]*SystemModuleVar{}
	mathModule.ModuleVarVec["pi"] = &SystemModuleVar{"pi", "The value of", "3.1415"}
	mathModule.ModuleVarVec["maxinteger"] = &SystemModuleVar{"maxinteger", "An integer with the maximum value for an integer", "number"}
	mathModule.ModuleVarVec["mininteger"] = &SystemModuleVar{"mininteger", "An integer with the minimum value for an integer", "number"}
	mathModule.ModuleVarVec["huge"] = &SystemModuleVar{"huge", "The float value `HUGE_VAL`, a value larger than any other numeric value", "number"}

	g.SystemModuleTipsMap["math"] = mathModule
	g.insertSysVarInfo("math", nil, &mathModule)

	// 4) os 模块
	osModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "os module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	osModule.ModuleFuncMap["clock"] = &SystemNoticeInfo{
		Detail:        "clock()",
		Documentation: "Returns an approximation of the amount in seconds of CPU time used by the program",
		FuncParamVec:  []FuncParamInfo{},
	}
	osModule.ModuleFuncMap["date"] = &SystemNoticeInfo{
		Detail:        "date(format, time)",
		Documentation: "Returns a string or a table containing date and time, formatted according to the given string `format`",
		FuncParamVec:  []FuncParamInfo{{"format", "format : string"}, {"time", "time : number"}},
	}
	osModule.ModuleFuncMap["difftime"] = &SystemNoticeInfo{
		Detail:        "difftime(t2, t1)",
		Documentation: "Returns the difference, in seconds, from time `t1` to time `t2`. (where the times are values returned by `os.time`). In POSIX, Windows, and some other systems, this value is exactly `t2`-`t1`",
		FuncParamVec:  []FuncParamInfo{{"t2", "t2: number"}, {"t1", "t1 : number"}},
	}

	osModule.ModuleFuncMap["execute"] = &SystemNoticeInfo{
		Detail:        "execute(command)",
		Documentation: " This function is equivalent to the C function `system`. It passes `command` to be executed by an operating system shell. Its first result is **true** if the command terminated successfully, or **nil** otherwise.",
		FuncParamVec:  []FuncParamInfo{{"command", "command : string"}},
	}
	osModule.ModuleFuncMap["exit"] = &SystemNoticeInfo{
		Detail:        "exit(code, close)",
		Documentation: "Calls the ISO C function `exit` to terminate the host program. If `code` is **true**, the returned status is `EXIT_SUCCESS`; if `code` is **false**, the returned status is `EXIT_FAILURE`; if `code` is a number, the returned status is this number. The default value for `code` is **true**",
		FuncParamVec:  []FuncParamInfo{{"code", "code : number"}, {"close", "close : bool"}},
	}

	osModule.ModuleFuncMap["getenv"] = &SystemNoticeInfo{
		Detail:        "getenv(varname)",
		Documentation: "Returns the value of the process environment variable `varname`, or **nil** if the variable is not defined",
		FuncParamVec:  []FuncParamInfo{{"varname", "varname : string"}},
	}

	osModule.ModuleFuncMap["remove"] = &SystemNoticeInfo{
		Detail:        "remove(filename)",
		Documentation: " Deletes the file (or empty directory, on POSIX systems) with the given name. If this function fails, it returns **nil**, plus a string describing the error and the error code. Otherwise, it returns true.",
		FuncParamVec:  []FuncParamInfo{{"filename", "filename: string"}},
	}
	osModule.ModuleFuncMap["rename"] = &SystemNoticeInfo{
		Detail:        "rename(oldname, newname)",
		Documentation: "Renames the file or directory named `oldname` to `newname`. If this function fails, it returns **nil**, plus a string describing the error and the error code. Otherwise, it returns true",
		FuncParamVec:  []FuncParamInfo{{"oldname", "oldname : string"}, {"newname", "newname : string"}},
	}
	osModule.ModuleFuncMap["setlocale"] = &SystemNoticeInfo{
		Detail:        "setlocale(locale, category)",
		Documentation: "Sets the current locale of the program. `locale` is a system-dependent string specifying a locale",
		FuncParamVec:  []FuncParamInfo{{"locale", "locale : string"}, {"category", "category : string"}},
	}
	osModule.ModuleFuncMap["time"] = &SystemNoticeInfo{
		Detail:        "time(table)",
		Documentation: "Returns the current time when called without arguments, or a time representing the date and time specified by the given table.",
		FuncParamVec:  []FuncParamInfo{{"table", "table : table"}},
	}
	osModule.ModuleFuncMap["tmpname"] = &SystemNoticeInfo{
		Detail:        "tmpname()",
		Documentation: "Returns a string with a file name that can be used for a temporary file. The file must be explicitly opened before its use and explicitly removed when no longer needed.",
		FuncParamVec:  []FuncParamInfo{},
	}
	g.SystemModuleTipsMap["os"] = osModule
	g.insertSysVarInfo("os", nil, &osModule)

	// 5) os 模块
	packageModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "package module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	packageModule.ModuleFuncMap["loadlib"] = &SystemNoticeInfo{
		Detail:        "loadlib(libname, funcname)",
		Documentation: "Dynamically links the host program with the C library `libname`",
		FuncParamVec:  []FuncParamInfo{{"libname", "libname : string"}, {"funcname", "funcname : string"}},
	}
	packageModule.ModuleFuncMap["searchpath"] = &SystemNoticeInfo{
		Detail:        "searchpath(name, path, sep, rep)",
		Documentation: "Searches for the given name in the given path",
		FuncParamVec:  []FuncParamInfo{{"name", "name : string"}, {"path", "path : string"}, {"sep", "sep : string"}, {"rep", "rep : string"}},
	}

	packageModule.ModuleVarVec = map[string]*SystemModuleVar{}
	packageModule.ModuleVarVec["config"] = &SystemModuleVar{"config", "", "A string describing some compile-time configurations for packages."}
	packageModule.ModuleVarVec["cpath"] = &SystemModuleVar{"cpath", "", "The path used by `require` to search for a C loader."}
	packageModule.ModuleVarVec["loaded"] = &SystemModuleVar{"loaded", "", "A table used by `require` to control which modules are already loaded"}
	packageModule.ModuleVarVec["path"] = &SystemModuleVar{"path", "", "The path used by `require` to search for a Lua loader"}
	packageModule.ModuleVarVec["preload"] = &SystemModuleVar{"preload", "", "A table to store loaders for specific modules (see `require`)"}
	packageModule.ModuleVarVec["searchers"] = &SystemModuleVar{"searchers", "", "A table used by require to control how to load modules"}
	g.SystemModuleTipsMap["package"] = packageModule
	g.insertSysVarInfo("package", nil, &packageModule)

	// 6) string 模块
	stringModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "string module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}
	stringModule.ModuleFuncMap["byte"] = &SystemNoticeInfo{
		Detail:        "byte(s, i, j)",
		Documentation: "Returns the internal numerical codes of the characters `s[i]`, `s[i+1]`,..., `s[j]",
		FuncParamVec:  []FuncParamInfo{{"s", "s :string"}, {"i", "i : number"}, {"j", "j : number"}},
	}
	stringModule.ModuleFuncMap["byte"] = &SystemNoticeInfo{
		Detail:        "char(...)",
		Documentation: "Receives zero or more integers. Returns a string with length equal to the number of arguments, in which each character has the internal numerical code equal to its corresponding argument",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}
	stringModule.ModuleFuncMap["dump"] = &SystemNoticeInfo{
		Detail:        "dump(func, strip)",
		Documentation: "Returns a string containing a binary representation (*a binary chunk*) of the given function, so that a later `load` on this string returns a copy of the function (but with new upvalues)",
		FuncParamVec:  []FuncParamInfo{{"func", "func : string"}, {"strip", "strip : string"}},
	}
	stringModule.ModuleFuncMap["find"] = &SystemNoticeInfo{
		Detail:        "find(s, pattern, init, plain)",
		Documentation: "Looks for the first match of `pattern` in the string `s`",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}, {"pattern", "pattern : string"}, {"init", "init : number"}, {"plain", "plain : string"}},
	}
	stringModule.ModuleFuncMap["format"] = &SystemNoticeInfo{
		Detail:        "format(formatstring, ...)",
		Documentation: "Returns a formatted version of its variable number of arguments following the description given in its first argument (which must be a string)",
		FuncParamVec:  []FuncParamInfo{{"formatstring", "formatstring : string"}, {"...", ""}},
	}
	stringModule.ModuleFuncMap["gmatch"] = &SystemNoticeInfo{
		Detail:        "gmatch(s, pattern)",
		Documentation: "Returns an iterator function that, each time it is called, returns the next captures from `pattern` over the string `s`. If `pattern` specifies no captures, then the whole match is produced in each call.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}, {"pattern", "pattern : string"}},
	}
	stringModule.ModuleFuncMap["gsub"] = &SystemNoticeInfo{
		Detail:        "gsub(s, pattern, repl, n)",
		Documentation: "Returns a copy of `s` in which all (or the first `n`, if given) occurrences of the `pattern` have been replaced by a replacement string specified by `repl`, which can be a string, a table, or a function.",
		FuncParamVec: []FuncParamInfo{{"s", "s : string"}, {"pattern", "pattern : string"},
			{"repl", "repl : string"}, {"n", "n : number"}},
	}
	stringModule.ModuleFuncMap["len"] = &SystemNoticeInfo{
		Detail:        "len(s) ",
		Documentation: "Receives a string and returns its length. The empty string `\"\"` has length 0.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}},
	}
	stringModule.ModuleFuncMap["lower"] = &SystemNoticeInfo{
		Detail:        "lower(s) ",
		Documentation: "Receives a string and returns a copy of this string with all uppercase letters changed to lowercase",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}},
	}
	stringModule.ModuleFuncMap["match"] = &SystemNoticeInfo{
		Detail:        "match(s, pattern, init)",
		Documentation: "Looks for the first *match* of `pattern` in the string `s`. If it finds one, then `match` returns the captures from the pattern; otherwise it returns **nil**.",
		FuncParamVec: []FuncParamInfo{{"s", "s : string"}, {"pattern", "pattern : string"},
			{"init", "init : number"}},
	}
	stringModule.ModuleFuncMap["pack"] = &SystemNoticeInfo{
		Detail:        "pack(fmt, v1, v2, ...)",
		Documentation: "Returns a binary string containing the values `v1`, `v2`, etc. packed (that is, serialized in binary form) according to the format string `fmt",
		FuncParamVec: []FuncParamInfo{{"fmt", "fmt : string"}, {"v1", "v1 ：string"},
			{"v2", "v2 : string"}, {"...", ""}},
	}
	stringModule.ModuleFuncMap["packsize"] = &SystemNoticeInfo{
		Detail:        "packsize(fmt)",
		Documentation: "Returns the size of a string resulting from `string.pack` with the given format. The format string cannot have the variable-length options '`s`' or '`z`'",
		FuncParamVec:  []FuncParamInfo{{"fmt", "fmt : string"}},
	}
	stringModule.ModuleFuncMap["rep"] = &SystemNoticeInfo{
		Detail:        "rep(s, n, sep)",
		Documentation: "Returns a string that is the concatenation of `n` copies of the string `s` separated by the string `sep`. The default value for `sep` is the empty string (that is, no separator). Returns the empty string if n is not positive",
		FuncParamVec: []FuncParamInfo{{"s", "s : string"}, {"n", "n : number"},
			{"sep", "sep : string"}},
	}

	stringModule.ModuleFuncMap["reverse"] = &SystemNoticeInfo{
		Detail:        "reverse(s)",
		Documentation: "Returns a string that is the string `s` reversed.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}},
	}

	stringModule.ModuleFuncMap["sub"] = &SystemNoticeInfo{
		Detail:        "sub(s, i, j)",
		Documentation: "Returns the substring of `s` that starts at `i` and continues until `j`; `i` and `j` can be negative.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}, {"i", "i : number"}, {"j", "j : number"}},
	}

	stringModule.ModuleFuncMap["unpack"] = &SystemNoticeInfo{
		Detail:        "unpack(fmt, s, pos)",
		Documentation: "Returns the values packed in string `s` according to the format string `fmt`. An optional `pos` marks where to start reading in `s` (default is 1). After the read values, this function also returns the index of the first unread byte in `s`",
		FuncParamVec:  []FuncParamInfo{{"fmt", "fmt : string"}, {"s", "s : string"}, {"pos", "pos : number"}},
	}

	stringModule.ModuleFuncMap["upper"] = &SystemNoticeInfo{
		Detail:        "upper(s) ",
		Documentation: "Receives a string and returns a copy of this string with all lowercase letters changed to uppercase",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}},
	}

	g.SystemModuleTipsMap["string"] = stringModule
	g.insertSysVarInfo("string", nil, &stringModule)

	// 7) table 模块
	tableModule := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "table module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}

	tableModule.ModuleFuncMap["concat"] = &SystemNoticeInfo{
		Detail:        "concat(list, sep, i, j)",
		Documentation: "Given a list where all elements are strings or numbers, returns the string `list[i]..sep..list[i+1] ... sep..list[j]`",
		FuncParamVec:  []FuncParamInfo{{"list", "list : table"}, {"sep", "sep : string"}, {"i", "i : number"}, {"j", "j : number"}},
	}

	tableModule.ModuleFuncMap["insert"] = &SystemNoticeInfo{
		Detail:        "insert(list, pos, value)",
		Documentation: "Inserts element `value` at position `pos` in `list`, shifting up the elements to `list[pos]`, `list[pos+1]`, `···`, `list[#list]`.",
		FuncParamVec:  []FuncParamInfo{{"list", "list : table"}, {"pos", "pos : number"}, {"value", "value : any"}},
	}

	tableModule.ModuleFuncMap["move"] = &SystemNoticeInfo{
		Detail:        "move(a1, f, e, t, a2)",
		Documentation: "Moves elements from table a1 to table `a2`, performing the equivalent to the following multiple assignment: `a2[t]`,`··· = a1[f]`,`···,a1[e]`. The default for `a2` is `a1`.",
		FuncParamVec: []FuncParamInfo{{"a1", "a1 : table"}, {"f", "f : number"},
			{"e", "e : number"}, {"t", "t : number"}, {"a2", "a2 : table"}},
	}

	tableModule.ModuleFuncMap["pack"] = &SystemNoticeInfo{
		Detail:        "pack(...)",
		Documentation: "Returns a new table with all arguments stored into keys 1, 2, etc. and with a field \"`n`\" with the total number of arguments",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}

	tableModule.ModuleFuncMap["remove"] = &SystemNoticeInfo{
		Detail:        "remove(list, pos)",
		Documentation: "Removes from `list` the element at position `pos`, returning the value of the removed element",
		FuncParamVec:  []FuncParamInfo{{"list", "list ：table"}, {"pos", "pos : number"}},
	}

	tableModule.ModuleFuncMap["sort"] = &SystemNoticeInfo{
		Detail:        "sort(list, comp)",
		Documentation: "Sorts list elements in a given order, *in-place*, from `list[1]` to `list[#list]`",
		FuncParamVec:  []FuncParamInfo{{"list", "list : table"}, {"comp", "comp : function"}},
	}

	tableModule.ModuleFuncMap["unpack"] = &SystemNoticeInfo{
		Detail:        "unpack(list, i, j)",
		Documentation: "Returns the elements from the given list. This function is equivalent to return `list[i]`, `list[i+1]`, `···`, `list[j]` By default, i is 1 and j is #list",
		FuncParamVec:  []FuncParamInfo{{"list", "list : table"}, {"i", "i : number"}, {"j", "j : number"}},
	}
	g.SystemModuleTipsMap["table"] = tableModule
	g.insertSysVarInfo("table", nil, &tableModule)

	// 8) utf8 模块
	utf8Module := OneModuleInfo{
		Detail:        "standard module",
		Documentation: "utf8 module",
		ModuleFuncMap: map[string]*SystemNoticeInfo{},
	}

	utf8Module.ModuleFuncMap["char"] = &SystemNoticeInfo{
		Detail:        "char(...)",
		Documentation: "Receives zero or more integers, converts each one to its corresponding UTF-8 byte sequence and returns a string with the concatenation of all these sequences.",
		FuncParamVec:  []FuncParamInfo{{"...", ""}},
	}

	utf8Module.ModuleFuncMap["codes"] = &SystemNoticeInfo{
		Detail:        "codes(s)",
		Documentation: " Returns values so that the construction > `for p, c in utf8.codes(s) do` *body* `end`.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : string"}},
	}

	utf8Module.ModuleFuncMap["codepoint"] = &SystemNoticeInfo{
		Detail:        "codepoint(s, i, j)",
		Documentation: "Returns the codepoints (as integers) from all characters in `s` that start between byte position `i` and `j` (both included). The default for `i` is 1  and for `j` is `i`. It raises an error if it meets any invalid byte sequence.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : table"}, {"i", "i : number"}, {"j", "j : number"}},
	}

	utf8Module.ModuleFuncMap["len"] = &SystemNoticeInfo{
		Detail:        "len(s, i, j)",
		Documentation: " Returns the number of UTF-8 characters in string `s` that start between positions `i` and `j` (both inclusive). The default for `i` is 1 and for `j` is -1. If it finds any invalid byte sequence, returns a false value plus the position of the first invalid byte.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : table"}, {"i", "i : number"}, {"j", "j : number"}},
	}

	utf8Module.ModuleFuncMap["len"] = &SystemNoticeInfo{
		Detail:        "len(s, n, i)",
		Documentation: "Returns the position (in bytes) where the encoding of the `n`-th character of `s` (counting from position `i`) starts.",
		FuncParamVec:  []FuncParamInfo{{"s", "s : table"}, {"n", "n : number"}, {"i", "i : number"}},
	}

	utf8Module.ModuleVarVec = map[string]*SystemModuleVar{}
	utf8Module.ModuleVarVec["charpattern"] = &SystemModuleVar{"charpattern", "",
		"The pattern (a string, not a function),which matches exactly one UTF-8 byte sequence, assuming that the subject is a valid UTF-8 string."}

	g.SystemModuleTipsMap["utf8"] = utf8Module
	g.insertSysVarInfo("utf8", nil, &utf8Module)
}

// GetSysVar 根据传入的变量，看是否在SysVarMap 里面
func (g *GlobalConfig) GetSysVar(strName string) *VarInfo {
	varInfo, _ := g.SysVarMap[strName]
	return varInfo
}

// IsModuleInsideVar 判断传入的是否为系统模块内的变量
func (g *GlobalConfig) IsModuleInsideVar(moduleName, keyName string) bool {
	varInfo, _ := g.SysVarMap[moduleName]
	if varInfo == nil {
		return false
	}

	if varInfo.SubMaps == nil {
		return false
	}

	subSymbol, _ := varInfo.SubMaps[keyName]
	if subSymbol == nil {
		return false
	}

	return true
}

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"luahelper-lsp/langserver/filefolder"
	"luahelper-lsp/langserver/log"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// SnippetItem 单个Snippet元素
type SnippetItem struct {
	InsertText string //  插入的文本
	Detail     string // 描述信息
}

// GlobalConfig 对外封装的全局配置信息
type GlobalConfig struct {
	// 是否读取了ylua.json配置文件，如果读取到了配置文件，获取相应的配置；如果没有读取到配置，默认以客户端的模式
	// 默认的客户端模式，会忽略一些告警，读取文件路径时候，尽量进行推导，可能存在同名的文件，尽量获取路径匹配度高的
	ReadJSONFlag bool

	// 如果是包含工程的入口文件，配置读取，后台专门定制的特性
	ProjectFiles []string

	// 是否开启告警
	showWarnFlag bool

	// 当一个lua文件引入其他的lua文件时候，需要匹配全路径，0为否，后台专业定制时候需要填写1
	ReferMatchPathFlag bool

	//lua内部定义的函数、模块、变量，检查时候进行忽略
	LuaInMap map[string]string

	// 主动忽略的变量、模块，配置文件中引入的
	IgnoreVarMap map[string]string

	// 主动忽略的通配符变量、模块，配置文件中引入的
	IgnoreWildcarVarMap []string

	// 主动引用其他的文件
	ReferOtherFileMap map[string]bool

	// 当引入文件时候，如果没有发现找不到对应的文件会进行报错，这里可以忽略哪些找不到的文件（如果能读取尽量进行读取）
	IgnoreReferFileMap map[string]bool

	// 忽略某个文件的中的未定义变量，第一层的map为文件lua名，第二层map为忽略的变量
	// 匹配的时候是包含关系，或是go的正则匹配
	IgnoreFileDefineVarMap map[string](map[string]bool)

	// 忽略指定文件中的指定类型错误, 第二层map为忽略的错误类型
	// 匹配的时候，包含文件名或go的正则
	IgnoreFileErrTypesMap map[string](map[int]bool)

	IgnoreFileErrTypesRegexp map[string]*regexp.Regexp

	// 全局忽略错误类型, 配置读取，如果没有配置，会默认设置一些值
	IgnoreErrorTypeMap map[CheckErrorType]bool

	// 开启的检查类型（白名单）, 配置读取，用于灰度测试
	OpenErrorTypeMap map[CheckErrorType]bool

	// 需要忽略分析的lua文件夹， 配置文件读取，包含文件名或go的正则
	IgnoreHandleFolderVec []string

	// 需要忽略分析的lua文件, 包含文件名或go的正则
	IgnoreHandleFileVec []string

	// 需要忽略告警的lua文件夹, 包含文件名或go的正则
	IgnoreErrorFloderVec []string

	// 需要忽略告警的lua文件，包含文件名或go的正则
	IgnoreErrorFileVec []string

	IgnoreErrorFileOrFloderRegexp map[string]*regexp.Regexp

	// 局部变量定义了，忽略这些
	IgnoreLocalNoUseVarMap map[string]bool

	// require("math") 时，忽略系统的系统的库
	IgnoreRequireSystemModule map[string]bool

	// 忽略与文件名同名的变量，特定客户端有这种需求
	// 这里的变量当所有的文件加载好后
	// 开始构建，去掉前缀和后缀， 只保留关键名 例如， one/two.lua 只保留two
	IgnoreSameFileNameMap map[string]bool

	//是否忽略与文件名同名的变量，客户端需要屏蔽这种变量
	// 默认为false，客户端专业设置的变量
	IgnoreFileNameVarFlag bool

	// 项目中特有的协议数组，例如有c2s, s2s，后台特有的
	ProtocolVars []string

	// 协议前缀变量未找到，是否告警, 默认告警, 默认告警，值为false
	ProtocolPreIngoreFlag bool

	// 代码补全时候，增加的提示关键字变量
	CompKeyMap map[string]bool

	// snippet代码补全
	CompSnippetMap map[string]SnippetItem

	// 忽略标准库中定义了为未使用的引用变量名
	ignoreSysNoUseMap map[string]bool

	// 忽略系统的注解类型，当为本地运行时候，忽略系统的
	ignoreSysAnnotateTypeMap map[string]bool

	// 查找文件是否存在的时候，用map进行cache，已经查找过的，存在的加快速度返回
	// 每次新进行扫描或保存的时候，清除掉缓存，防止有cache导致失败
	FileExistCacheMap map[string]bool

	// 文件是否存在的互斥锁
	FileExistCacheMutex sync.Mutex

	// 框架中新引入的文件形式
	ReferFrameFiles []referFrameFile

	// 项目中引入其他文件，路径分隔符，默认为. 例如require("one.b") 表示引入one/b.lua 文件
	PathSeparator string

	// 引入另外一个目录，可以用于设置引入额外LuaHelper注解格式文件夹
	OtherDir string

	// 所有后缀文件关联到lua类型, 例如有的.txt后缀文件，会当成lua文件处理
	AssocialList []string

	ReferenceMaxNum     int  // 查找引用时候，返回的最大的引用数量
	ReferenceDefineFlag bool // 查找引用时候，是否需要显示定义
	PreviewFieldsNum    int  // 当hover一个table时，显示最多field的数量

	// 查询_G.a 这样的全局符号，a是否会扩大到全局符号定义
	// 例如前面定义了a=1,  那么此时_G.a 会指向前面的a=1
	GVarExtendGlobalFlag bool

	// 是否区分 ：与 . 成员的调用标记。0表示.可以调用:用法；1表示两者不能相互调用；2表示两者可以相互调用
	colonFlag int

	// 配置的注解配置
	anntotateSets []AnntotateSet

	// 所有的目录管理
	dirManager *DirManager

	// 当为emacs或vim时增加系统库的代码补全提示
	SystemTipsMap map[string]SystemNoticeInfo

	// 标准模块库模块的提升
	SystemModuleTipsMap map[string]OneModuleInfo

	// 存放所有标准库和模块的全局变量，map管理；系统的函数和模块转换成想要的VarInfo，统一起来
	SysVarMap map[string]*VarInfo
}

// GConfig *GlobalConfig 全局配置对象初始化
var GConfig *GlobalConfig

// 创建默认的GlobalConfig
func createDefaultGlobalConfig() {
	GConfig = &GlobalConfig{
		ReadJSONFlag:           false,
		ReferMatchPathFlag:     false,
		showWarnFlag:           false,
		ReferenceMaxNum:        3000,
		PreviewFieldsNum:       30,
		ReferenceDefineFlag:    true,
		GVarExtendGlobalFlag:   true,
		colonFlag:              0,
		IgnoreFileNameVarFlag:  false,
		ProtocolVars:           []string{},
		ProtocolPreIngoreFlag:  false,
		ReferOtherFileMap:      map[string]bool{},
		IgnoreLocalNoUseVarMap: map[string]bool{},
		IgnoreWildcarVarMap:    []string{},
		PathSeparator:          ".",
		anntotateSets:          []AnntotateSet{},
		dirManager:             createDirManager(),
		OtherDir:               "",
	}
}

// OneTipFile 提示的一个文件信息
type OneTipFile struct {
	StrName string // 提示的内容
	StrDesc string // 其他的描述
}

// json配置文件的定义
type (
	ignoreFileVar struct {
		Name string   `json:"File"`
		Vars []string `json:"Vars"`
	}

	ignoreFileErrType struct {
		Name  string `json:"File"`
		Types []int  `json:"Types"`
	}

	// 解决新的引入一个文件的方式，传统的引入文件的方式是require，一些框架可能类似require的功能
	// 通常的框架，例如hive框架会使用import函数导入一个文件，也会类似require的功能
	referFrameFile struct {
		// 导入函数的名称，例如为import
		Name string `json:"Name"`

		// 引入的类型：0是类似import；1是类似require；2是两种都有可能，自动选择看文件的返回，如果返回table就是require，否则是import
		Type int `json:"type"`

		// 导入函数是否要包括文件名后缀，包含后缀的为 import("one.lua"), 不含后缀的为import("one")
		SuffixFlag int `json:"SuffixFlag"`
	}

	// AnntotateSet 注解推导的方式，配置可以标明
	AnntotateSet struct {
		// 函数的名称，哪些函数能够尝试自动关联注解
		FuncName string `json:"FuncName"`

		// 函数哪个一个参数有注解推导相关，默认从1开始
		ParamIndex int `json:"ParamIndex"`

		// 是否关联第二层函数的返回
		//SecondFuncName string `json:"SecondFuncName"`

		// 关联的第二层函数的哪一个参数，默认从1开始
		//SecondParamIndex int `json:"SecondParamIndex"`

		// 函数的参数是否需要拆分，1为需要，0为不要，当为需要的时候 以.和 / 分割，最后一个词
		// 例如 /Game/Mocy/BPMocyWidgetChild.BPMocyWidgetChild_C ，配置为1的时候，为BPMocyWidgetChild_C
		SplitFlag int `json:"SplitFlag"`

		// 注解类型为前面推导出来的字符串增加下面的前缀，默认为空
		PrefixStr string `json:"PrefixStr"`

		// 注解类型为前面推导出来的字符串增加下面的前缀，为可选的类别之一
		PrefixStrList []string `json:"PrefixStrList"`

		// 注解类型为前面推导出来的字符串增加下面的后缀，默认为空
		SuffixStr string `json:"SuffixStr"`
	}

	// JSONConfig 对外封装的json全局配置信息
	JSONConfig struct {
		BaseDir               string              `json:"BaseDir"`               // 所有工程的根目录
		ShowWarnFlag          int                 `json:"ShowWarnFlag"`          // 是否显示告警, 0为不显示，其他的为显示
		ReferMatchPathFlag    int                 `json:"ReferMatchPathFlag"`    // 当一个lua文件引入其他的lua文件时候，需要匹配全路径，0为否
		IgnoreFileNameVarFlag int                 `json:"IgnoreFileNameVarFlag"` // 是否忽略与文件名同名的变量，客户端需要屏蔽这种变量
		ProjectFiles          []string            `json:"ProjectFiles"`          // 工程的入口文件
		IgnoreModules         []string            `json:"IgnoreModules"`         // 忽略的模块
		IgnoreWildcardModules []string            `json:"IgnoreWildcardModules"` // 忽略的包含通配符的模块
		IgnoreFileVars        []ignoreFileVar     `json:"IgnoreFileVars"`        // 忽略指定文件中的变量
		IgnoreReadFiles       []string            `json:"IgnoreReadFiles"`       // 读不到某些文件时候，不报错，忽略（windows可能没有某些配置文件）
		IgnoreErrorTypes      []int               `json:"IgnoreErrorTypes"`      // 忽略指定类型的错误
		IgnoreFileOrFloder    []string            `json:"IgnoreFileOrFloder"`    // 忽略分析的文件或文件夹
		IgnoreFileErr         []string            `json:"IgnoreFileErr"`         // 忽略下列文件中的错误
		IgnoreFileErrTypes    []ignoreFileErrType `json:"IgnoreFileErrTypes"`    // 忽略指定文件中的指定类型错误
		IgnoreLocalNoUseVars  []string            `json:"IgnoreLocalNoUseVars"`  // 忽略哪些局部变量定义了未使用的
		ProtocolVars          []string            `json:"ProtocolVars"`          // 项目中特有的协议数组，例如有c2s, s2s
		ProtocolPreIngoreFlag int                 `json:"ProtocolPreIngoreFlag"` // 协议前缀变量未找到，是否告警, 默认告警
		ReferFrameFiles       []referFrameFile    `json:"ReferFrameFiles"`       // 项目中引用其他的框架文件
		PathSeparator         string              `json:"PathSeparator"`         // 项目中引入其他文件，路径分隔符，默认为. 例如require("one.b") 表示引入one/b.lua 文件
		AnntotateSets         []AnntotateSet      `json:"AnntotateSets"`         // 自动推导的注解方式
		OtherDir              string              `json:"OtherDir"`              // 引入另外一个目录，可以用于设置引入额外LuaHelper注解格式文件夹
		OpenErrorTypes        []int               `json:"OpenErrorTypes"`        // 开启的告警项
	}
)

// 对外封装的json全局配置信息
var jsonConfig *JSONConfig

// 创建默认的json config
func createDefaultJSONCfig() {
	jsonConfig = &JSONConfig{
		BaseDir:               "./",
		ShowWarnFlag:          1,
		ReferMatchPathFlag:    0,
		IgnoreFileNameVarFlag: 0,
		ProjectFiles:          []string{},
		IgnoreModules:         []string{},
		IgnoreWildcardModules: []string{},
		IgnoreFileVars:        []ignoreFileVar{},
		IgnoreReadFiles:       []string{},
		IgnoreErrorTypes:      []int{},
		IgnoreFileOrFloder:    []string{},
		IgnoreFileErr:         []string{},
		IgnoreFileErrTypes:    []ignoreFileErrType{},
		IgnoreLocalNoUseVars:  []string{},
		ProtocolVars:          []string{},
		ProtocolPreIngoreFlag: 0,
		ReferFrameFiles:       []referFrameFile{{Name: "import", Type: 0, SuffixFlag: 1}},
		PathSeparator:         ".",
		AnntotateSets:         []AnntotateSet{},
		OpenErrorTypes:        []int{},
	}
}

// GlobalConfigDefautInit 默认的初始化
func GlobalConfigDefautInit() {
	// 创建默认的GlobalConfig
	createDefaultGlobalConfig()

	// 创建默认的json config
	createDefaultJSONCfig()
}

func (g *GlobalConfig) setSysNotUseMap() {
	g.ignoreSysNoUseMap = map[string]bool{}
	g.ignoreSysNoUseMap["assert"] = true
	g.ignoreSysNoUseMap["collectgarbage"] = true
	g.ignoreSysNoUseMap["dofile"] = true
	g.ignoreSysNoUseMap["error"] = true
	g.ignoreSysNoUseMap["getmetatable"] = true
	g.ignoreSysNoUseMap["ipairs"] = true
	g.ignoreSysNoUseMap["load"] = true
	g.ignoreSysNoUseMap["loadfile"] = true
	g.ignoreSysNoUseMap["next"] = true
	g.ignoreSysNoUseMap["pairs"] = true
	g.ignoreSysNoUseMap["pcall"] = true
	g.ignoreSysNoUseMap["print"] = true
	g.ignoreSysNoUseMap["rawequal"] = true
	g.ignoreSysNoUseMap["rawget"] = true
	g.ignoreSysNoUseMap["rawlen"] = true
	g.ignoreSysNoUseMap["rawset"] = true
	g.ignoreSysNoUseMap["require"] = true
	g.ignoreSysNoUseMap["select"] = true
	g.ignoreSysNoUseMap["setmetatable"] = true
	g.ignoreSysNoUseMap["tonumber"] = true
	g.ignoreSysNoUseMap["tostring"] = true
	g.ignoreSysNoUseMap["type"] = true
	g.ignoreSysNoUseMap["xpcall"] = true
	g.ignoreSysNoUseMap["coroutine"] = true
	g.ignoreSysNoUseMap["debug"] = true
	g.ignoreSysNoUseMap["io"] = true
	g.ignoreSysNoUseMap["file"] = true
	g.ignoreSysNoUseMap["math"] = true
	g.ignoreSysNoUseMap["os"] = true
	g.ignoreSysNoUseMap["package"] = true
	g.ignoreSysNoUseMap["string"] = true
	g.ignoreSysNoUseMap["table"] = true
	g.ignoreSysNoUseMap["utf8"] = true
}

// IntialGlobalVar 初始化全局配置
func (g *GlobalConfig) IntialGlobalVar() {
	g.LuaInMap = map[string]string{
		"_VERSION":       "var",
		"_ENV":           "var",
		"_env":           "var",
		"self":           "var_class",
		"assert":         "function",
		"collectgarbage": "function",
		"dofile":         "function",
		"error":          "function",
		"getmetatable":   "function",
		"ipairs":         "function",
		"load":           "function",
		"loadfile":       "function",
		"next":           "function",
		"pairs":          "function",
		"pcall":          "function",
		"print":          "function",
		"rawequal":       "function",
		"rawget":         "function",
		"rawlen":         "function",
		"rawset":         "function",
		"require":        "function",
		"select":         "function",
		"setmetatable":   "function",
		"tonumber":       "function",
		"tostring":       "function",
		"type":           "function",
		"xpcall":         "function",
		"loadstring":     "function",
		"import":         "function",
		"_G":             "module",
		"coroutine":      "module",
		"debug":          "module",
		"io":             "module",
		"file":           "module",
		"math":           "module",
		"os":             "module",
		"package":        "module",
		"string":         "module",
		"table":          "module",
		"utf8":           "module",
	}

	g.ReferOtherFileMap = map[string]bool{
		"import":  true,
		"dofile":  true,
		"require": true,
	}

	g.IgnoreErrorTypeMap = map[CheckErrorType]bool{}
	g.OpenErrorTypeMap = map[CheckErrorType]bool{}

	g.FileExistCacheMap = map[string]bool{}
	g.CompKeyMap = map[string]bool{}

	g.CompKeyMap["and"] = true
	g.CompKeyMap["break"] = true
	g.CompKeyMap["::"] = true

	g.CompKeyMap["end"] = true
	g.CompKeyMap["goto"] = true
	g.CompKeyMap["in"] = true
	g.CompKeyMap["local"] = true
	g.CompKeyMap["not"] = true
	g.CompKeyMap["or"] = true
	g.CompKeyMap["return"] = true
	g.CompKeyMap["until"] = true
	g.CompKeyMap["false"] = true
	g.CompKeyMap["nil"] = true
	g.CompKeyMap["true"] = true
	g.CompKeyMap["self"] = true
	g.CompKeyMap["continue"] = true
	g.CompKeyMap["_G"] = true
	g.CompKeyMap["local"] = true

	g.CompSnippetMap = map[string]SnippetItem{}
	g.CompSnippetMap["else"] = SnippetItem{
		InsertText: "else\n\t",
		Detail:     "else\n\t",
	}

	g.CompSnippetMap["for .. ipairs"] = SnippetItem{
		InsertText: "for ${1:i}, ${2:v} in ipairs(${3:t}) do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for i, v in ipairs(t) do" + "\n\n" + "end",
	}

	g.CompSnippetMap["for .. pairs"] = SnippetItem{
		InsertText: "for ${1:k}, ${2:v} in pairs(${3:t}) do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for k, v in pairs(t) do" + "\n\n" + "end",
	}

	g.CompSnippetMap["then .. end"] = SnippetItem{
		InsertText: "then" + "\n" + "\t" + "${0:}" + "\n" + "end",
		Detail:     "then" + "\n\n" + "end",
	}

	g.CompSnippetMap["for i = .."] = SnippetItem{
		InsertText: "for ${1:i} = ${2:1}, ${3:10}, ${4:1} do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for i = 1, 10, 1 do" + "\n\n" + "end",
	}

	g.CompSnippetMap["if"] = SnippetItem{
		InsertText: "if ${1:condition} then\n\t${0:}\nend",
		Detail:     "if condition then\n\nend",
	}

	g.CompSnippetMap["elseif"] = SnippetItem{
		InsertText: "elseif ${1:condition} then\n\t${0:}",
		Detail:     "elseif condition then ..",
	}

	g.CompSnippetMap["for"] = SnippetItem{
		InsertText: "for ${1:k}, ${2:v} in pairs(${3:t}) do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for k, v in pairs(t) do" + "\n\n" + "end",
	}

	g.CompSnippetMap["fori"] = SnippetItem{
		InsertText: "for ${1:i}, ${2:v} in ipairs(${3:t}) do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for i, v in ipairs(t) do" + "\n\n" + "end",
	}

	g.CompSnippetMap["forp"] = SnippetItem{
		InsertText: "for ${1:k}, ${2:v} in pairs(${3:t}) do" + "\n\t" + "$0" + "\n" + "end",
		Detail:     "for k, v in pairs(t) do" + "\n\n" + "end",
	}

	g.CompSnippetMap["do"] = SnippetItem{
		InsertText: "do\n${0:}\nend",
		Detail:     "do\n\nend",
	}

	g.CompSnippetMap["while"] = SnippetItem{
		InsertText: "while ${1:condition} do\n\t${0:}\nend",
		Detail:     "while\n\nend",
	}

	g.CompSnippetMap["repeat"] = SnippetItem{
		InsertText: "repeat\n\t${0:}\nuntil ${1:condition}",
		Detail:     "repeat\n\nuntil",
	}

	g.CompSnippetMap["local function"] = SnippetItem{
		InsertText: "local function ${1:func}(${2:})\n\t${0:}\nend",
		Detail:     "local function func()\n\nend",
	}

	g.CompSnippetMap["function"] = SnippetItem{
		InsertText: "function ${1:func}(${2:})\n\t${0:}\nend",
		Detail:     "function func()\n\nend",
	}

	// 设置系统忽略定义未使用的变量
	g.setSysNotUseMap()

	// 设置系统库的代码补全提示
	g.InitSystemTips()

	// 忽略系统的require 模块
	g.IgnoreRequireSystemModule = map[string]bool{
		"table":       true,
		"string":      true,
		"coroutine":   true,
		"io":          true,
		"os":          true,
		"math":        true,
		"socket.core": true,
		"debug":       true,
	}

	g.IgnoreSameFileNameMap = map[string]bool{}
	g.ignoreSysAnnotateTypeMap = map[string]bool{}
}

// HandleNotJSONCheckFlag 由于没有读取到配置，忽略一些特定的告警
// ignoreFileOrDir 为工程忽略指定检查的文件和文件夹
// ignoreFileOrDirErr 为工程忽略指定文件和文件夹的错误
func (g *GlobalConfig) handleNotJSONCheckFlag(checkFlagList []bool, ignoreFileOrDir []string, ignoreFileOrDirErr []string) {
	// 没有读取到了json文件
	g.ReadJSONFlag = false
	g.OtherDir = ""

	// 添加引入文件的方式
	for _, oneReferFrame := range jsonConfig.ReferFrameFiles {
		GConfig.ReferOtherFileMap[oneReferFrame.Name] = true
		GConfig.LuaInMap[oneReferFrame.Name] = "function"
	}
	GConfig.ReferFrameFiles = jsonConfig.ReferFrameFiles

	// 忽略对某些文件或文件夹进行check分析， 包含go语言的正则
	g.IgnoreHandleFolderVec = make([]string, 0, 2)
	g.IgnoreHandleFileVec = make([]string, 0, 2)
	for _, fileOrFloder := range ignoreFileOrDir {
		if strings.HasSuffix(fileOrFloder, ".lua") {
			g.IgnoreHandleFileVec = append(g.IgnoreHandleFileVec, fileOrFloder)
		} else {
			g.IgnoreHandleFolderVec = append(g.IgnoreHandleFolderVec, fileOrFloder)
		}
	}

	// 忽略告警的文件和文件夹
	g.IgnoreErrorFloderVec = make([]string, 0, 2)
	g.IgnoreErrorFileVec = make([]string, 0, 2)
	g.IgnoreErrorFileOrFloderRegexp = map[string]*regexp.Regexp{}
	for _, fileOrFloder := range ignoreFileOrDirErr {
		if strings.HasSuffix(fileOrFloder, ".lua") {
			g.IgnoreErrorFileVec = append(g.IgnoreErrorFileVec, fileOrFloder)
		} else {
			g.IgnoreErrorFloderVec = append(g.IgnoreErrorFloderVec, fileOrFloder)
		}
		g.IgnoreErrorFileOrFloderRegexp[fileOrFloder] = regexp.MustCompile(fileOrFloder)
	}
	// 忽略客户端额外的Lua文件夹告警
	g.IgnoreErrorFloderVec = append(g.IgnoreErrorFloderVec, "server/meta")
	g.IgnoreErrorFileOrFloderRegexp["server/meta"] = regexp.MustCompile("server/meta")

	listLen := len(checkFlagList)
	if listLen < 1 {
		return
	}

	// 是否全部屏蔽
	if !checkFlagList[0] {
		g.showWarnFlag = false
		for i := CheckErrorSyntax; i < CheckErrorMax; i++ {
			g.IgnoreErrorTypeMap[(CheckErrorType)(i)] = true
		}
		return
	}

	g.showWarnFlag = true
	g.IgnoreErrorTypeMap = map[CheckErrorType]bool{}
	for i := CheckErrorSyntax; i < CheckErrorMax; i++ {
		if i > listLen-1 {
			g.IgnoreErrorTypeMap[(CheckErrorType)(i)] = true
		} else {
			oneFlag := checkFlagList[i]
			if !oneFlag {
				g.IgnoreErrorTypeMap[(CheckErrorType)(i)] = true
			}
		}
	}

	g.IgnoreVarMap = map[string]string{}
}

// HandleChangeCheckList 客户端的告警配置有改动
func (g *GlobalConfig) HandleChangeCheckList(checkFlagList []bool, ignoreFileOrDir []string, ignoreFileOrDirErr []string) {
	if g.ReadJSONFlag {
		// 如果是读取json配置，忽略客户端的配置
		return
	}

	g.handleNotJSONCheckFlag(checkFlagList, ignoreFileOrDir, ignoreFileOrDirErr)
}

// ReadConfig strDir为配置文件的根目录
// configFileName 为配置文件的名称
// 返回值为空表示成功，其他表示失败
func (g *GlobalConfig) ReadConfig(strDir, configFileName string, checkFlagList []bool, ignoreFileOrDir []string,
	ignoreFileOrDirErr []string) error {
	strPath := g.dirManager.GetCompletePath(strDir, configFileName)

	bytes, err := ioutil.ReadFile(strPath)
	if err != nil {
		log.Debug("not find %s file", configFileName)
		// 没有读取到配置文件，设置一些默认值，忽略特定的告警
		g.handleNotJSONCheckFlag(checkFlagList, ignoreFileOrDir, ignoreFileOrDirErr)
		return nil
	}

	//var data configInfo
	err = json.Unmarshal(bytes, jsonConfig)
	if err != nil {
		strErr := fmt.Sprintf("read %s error=%s, json format error", configFileName, err.Error())
		return errors.New(strErr)
	}

	g.anntotateSets = jsonConfig.AnntotateSets

	// 读取到了json文件
	g.ReadJSONFlag = true

	if jsonConfig.BaseDir == "" {
		jsonConfig.BaseDir = "./"
	}

	g.OtherDir = jsonConfig.OtherDir
	g.dirManager.setConfigRelativeDir(jsonConfig.BaseDir)

	g.ProjectFiles = jsonConfig.ProjectFiles

	g.ReferMatchPathFlag = (jsonConfig.ReferMatchPathFlag == 1)
	g.showWarnFlag = (jsonConfig.ShowWarnFlag == 1)

	g.IgnoreFileNameVarFlag = (jsonConfig.IgnoreFileNameVarFlag == 1)

	g.ProtocolVars = jsonConfig.ProtocolVars
	g.ProtocolPreIngoreFlag = false
	if jsonConfig.ProtocolPreIngoreFlag == 1 {
		g.ProtocolPreIngoreFlag = true
	}

	// 忽略的模块
	g.IgnoreVarMap = map[string]string{}
	for _, modeStr := range jsonConfig.IgnoreModules {
		g.IgnoreVarMap[modeStr] = "module"
	}

	g.IgnoreWildcarVarMap = []string{}
	g.IgnoreWildcarVarMap = append(g.IgnoreWildcarVarMap, jsonConfig.IgnoreWildcardModules...)

	// 指定文件，忽略的变量
	g.IgnoreFileDefineVarMap = map[string](map[string]bool){}
	for _, ignoreFileVars := range jsonConfig.IgnoreFileVars {
		fileMapVars := map[string]bool{}
		for _, fileVarsStr := range ignoreFileVars.Vars {
			fileMapVars[fileVarsStr] = true
		}

		g.IgnoreFileDefineVarMap[ignoreFileVars.Name] = fileMapVars
	}

	// 忽略某些读不到的文件，不进行报错
	g.IgnoreReferFileMap = map[string]bool{}
	for _, ignoreFileStr := range jsonConfig.IgnoreReadFiles {
		g.IgnoreReferFileMap[ignoreFileStr] = true
	}

	// 忽略指定类型的错误
	g.IgnoreErrorTypeMap = map[CheckErrorType]bool{}
	for _, errorType := range jsonConfig.IgnoreErrorTypes {
		g.IgnoreErrorTypeMap[(CheckErrorType)(errorType)] = true
	}

	// 开启指定类型的告警检查
	g.OpenErrorTypeMap = map[CheckErrorType]bool{}
	for _, errorType := range jsonConfig.OpenErrorTypes {
		g.OpenErrorTypeMap[(CheckErrorType)(errorType)] = true
	}

	// 忽略对某些文件或文件夹进行check分析， 包含go语言的正则
	g.IgnoreHandleFolderVec = make([]string, 0, 2)
	g.IgnoreHandleFileVec = make([]string, 0, 2)
	for _, fileOrFloder := range jsonConfig.IgnoreFileOrFloder {
		if strings.HasSuffix(fileOrFloder, ".lua") {
			g.IgnoreHandleFileVec = append(g.IgnoreHandleFileVec, fileOrFloder)
		} else {
			g.IgnoreHandleFolderVec = append(g.IgnoreHandleFolderVec, fileOrFloder)
		}
	}

	// 忽略指定文件中的指定错误
	g.IgnoreFileErrTypesMap = map[string](map[int]bool){}
	g.IgnoreFileErrTypesRegexp = map[string]*regexp.Regexp{}
	for _, ignoreTypeVars := range jsonConfig.IgnoreFileErrTypes {
		fileTypesVars := map[int]bool{}
		for _, typeError := range ignoreTypeVars.Types {
			fileTypesVars[typeError] = true
		}

		g.IgnoreFileErrTypesMap[ignoreTypeVars.Name] = fileTypesVars
		g.IgnoreFileErrTypesRegexp[ignoreTypeVars.Name] = regexp.MustCompile(ignoreTypeVars.Name)
	}

	// 忽略告警的文件和文件夹
	g.IgnoreErrorFloderVec = make([]string, 0, 2)
	g.IgnoreErrorFileVec = make([]string, 0, 2)
	g.IgnoreErrorFileOrFloderRegexp = map[string]*regexp.Regexp{}
	for _, fileOrFloder := range jsonConfig.IgnoreFileErr {
		if strings.HasSuffix(fileOrFloder, ".lua") {
			g.IgnoreErrorFileVec = append(g.IgnoreErrorFileVec, fileOrFloder)
		} else {
			g.IgnoreErrorFloderVec = append(g.IgnoreErrorFloderVec, fileOrFloder)
		}
		g.IgnoreErrorFileOrFloderRegexp[fileOrFloder] = regexp.MustCompile(fileOrFloder)
	}
	// 忽略客户端额外的Lua文件夹告警
	g.IgnoreErrorFloderVec = append(g.IgnoreErrorFloderVec, "server/meta")
	g.IgnoreErrorFileOrFloderRegexp["server/meta"] = regexp.MustCompile("server/meta")

	// 局部变量定义了，未使用，忽略
	g.IgnoreLocalNoUseVarMap = map[string]bool{}
	for _, noUseStr := range jsonConfig.IgnoreLocalNoUseVars {
		g.IgnoreLocalNoUseVarMap[noUseStr] = true
	}

	// 默认为后台的hive框架，会import引入一个文件，并且包含后缀，引入的方式为 import("one.lua")
	if len(jsonConfig.ReferFrameFiles) == 0 {
		jsonConfig.ReferFrameFiles = []referFrameFile{{Name: "import", Type: 0, SuffixFlag: 1}}
	}

	// 添加引入文件的方式
	for _, oneReferFrame := range jsonConfig.ReferFrameFiles {
		g.ReferOtherFileMap[oneReferFrame.Name] = true
		g.LuaInMap[oneReferFrame.Name] = "function"
	}
	GConfig.ReferFrameFiles = jsonConfig.ReferFrameFiles

	if jsonConfig.PathSeparator != "" {
		GConfig.PathSeparator = jsonConfig.PathSeparator
	}

	log.Debug("read ok")
	return nil
}

// SetRequirePathSeparator 设置require其他lua文件时候的路径分割符
func (g *GlobalConfig) SetRequirePathSeparator(pathSeparator string) {
	if g.ReadJSONFlag {
		// 如果是读取json配置文件，忽略设置
		return
	}

	if pathSeparator != "." && pathSeparator != "/" {
		pathSeparator = "."
	}

	GConfig.PathSeparator = pathSeparator
}

// SetPreviewFieldsNum set preview fields num
func (g *GlobalConfig) SetPreviewFieldsNum(num int) {
	if num > 0 {
		GConfig.PreviewFieldsNum = num
	}
}

// InsertIngoreSystemModule 如果为本地形式运行，加载不了插件前端的Lua额外文件夹，忽略系统模块。批量插入
func (g *GlobalConfig) InsertIngoreSystemModule() {
	g.IgnoreVarMap["debug"] = "module"
	g.IgnoreVarMap["math"] = "module"
	g.IgnoreVarMap["os"] = "module"
	g.IgnoreVarMap["io"] = "module"
	g.IgnoreVarMap["coroutine"] = "module"
	g.IgnoreVarMap["utf8"] = "module"
	g.IgnoreVarMap["table"] = "module"
	g.IgnoreVarMap["string"] = "module"
	g.IgnoreVarMap["package"] = "module"
	g.IgnoreVarMap["bit"] = "module"
	g.IgnoreVarMap["bit32"] = "module"
	g.IgnoreVarMap["jit"] = "module"
	g.IgnoreVarMap["arg"] = "variable"
	g.IgnoreVarMap["_G"] = "variable"
	g.IgnoreVarMap["_VERSION"] = "variable"
	g.IgnoreVarMap["assert"] = "function"
	g.IgnoreVarMap["collectgarbage"] = "function"
	g.IgnoreVarMap["dofile"] = "function"
	g.IgnoreVarMap["error"] = "function"
	g.IgnoreVarMap["getfenv"] = "function"
	g.IgnoreVarMap["getmetatable"] = "function"
	g.IgnoreVarMap["ipairs"] = "function"
	g.IgnoreVarMap["load"] = "function"
	g.IgnoreVarMap["loadfile"] = "function"
	g.IgnoreVarMap["loadstring"] = "function"
	g.IgnoreVarMap["module"] = "function"
	g.IgnoreVarMap["next"] = "function"
	g.IgnoreVarMap["pairs"] = "function"
	g.IgnoreVarMap["pcall"] = "function"
	g.IgnoreVarMap["print"] = "function"
	g.IgnoreVarMap["rawequal"] = "function"
	g.IgnoreVarMap["rawget"] = "function"
	g.IgnoreVarMap["rawlen"] = "function"
	g.IgnoreVarMap["rawset"] = "function"
	g.IgnoreVarMap["require"] = "function"
	g.IgnoreVarMap["select"] = "function"
	g.IgnoreVarMap["setfenv"] = "function"
	g.IgnoreVarMap["setmetatable"] = "function"
	g.IgnoreVarMap["tonumber"] = "function"
	g.IgnoreVarMap["tostring"] = "function"
	g.IgnoreVarMap["type"] = "function"
	g.IgnoreVarMap["warn"] = "function"
	g.IgnoreVarMap["xpcall"] = "function"
	g.IgnoreVarMap["unpack"] = "function"
	g.IgnoreVarMap["require"] = "function"
}

// InsertIngoreSystemAnnotateType 当为本地运行时，忽略系统的注解类型type。批量插入
func (g *GlobalConfig) InsertIngoreSystemAnnotateType() {
	g.ignoreSysAnnotateTypeMap["any"] = true
	g.ignoreSysAnnotateTypeMap["nil"] = true
	g.ignoreSysAnnotateTypeMap["boolean"] = true
	g.ignoreSysAnnotateTypeMap["number"] = true
	g.ignoreSysAnnotateTypeMap["integer"] = true
	g.ignoreSysAnnotateTypeMap["thread"] = true
	g.ignoreSysAnnotateTypeMap["table"] = true
	g.ignoreSysAnnotateTypeMap["string"] = true
	g.ignoreSysAnnotateTypeMap["void"] = true
	g.ignoreSysAnnotateTypeMap["userdata"] = true
	g.ignoreSysAnnotateTypeMap["lightuserdata"] = true
	g.ignoreSysAnnotateTypeMap["function"] = true
}

// IsDefaultAnnotateType 判断是否为系统默认的注解类型
func (g *GlobalConfig) IsDefaultAnnotateType(str string) bool {
	_, ok := g.ignoreSysAnnotateTypeMap[str]
	return ok
}

// IsIngoreNeedReadFile 判断是否为忽略一定要读到的文件（有些配置文件，在windows端不会生成，此时windows读不到时不报错）
func (g *GlobalConfig) IsIngoreNeedReadFile(str string) bool {
	return g.IgnoreReferFileMap[str]
}

// IsIgnoreFileDefineVar 判断是否为忽略的文件中的定义的变量
func (g *GlobalConfig) IsIgnoreFileDefineVar(luaFile string, strName string) bool {
	// todo 这个函数被很多地方调用，是否有优化到空间
	for fileName, sencondMap := range g.IgnoreFileDefineVarMap {
		if strings.Contains(luaFile, fileName) {
			if _, ok := sencondMap[strName]; ok {
				return true
			}
		}
	}

	return false
}

// IsIgnoreErrorFile 判断文件是否为需要忽略告警的
func (g *GlobalConfig) IsIgnoreErrorFile(strFile string, errType CheckErrorType) bool {
	if !g.showWarnFlag {
		return true
	}

	if _, ok := g.IgnoreErrorTypeMap[errType]; ok {
		return true
	}

	for _, floderStr := range g.IgnoreErrorFloderVec {
		if strings.Contains(strFile, floderStr) {
			log.Debug("path=%s is ignore handle err, mode=%s", strFile, floderStr)
			return true
		}

		oneRegex, ok := g.IgnoreErrorFileOrFloderRegexp[floderStr]
		if !ok {
			continue
		}

		if oneRegex.MatchString(strFile) {
			log.Debug("path=%s is ignore handle floder err, mode=%s", floderStr, strFile)
			return true
		}
		// if ok, _ := regexp.Match(floderStr, []byte(strFile)); ok {
		// 	log.Debug("path=%s is ignore handle floder err, mode=%s", floderStr, strFile)
		// 	return true
		// }
	}

	for _, fileStr := range g.IgnoreErrorFileVec {
		if strings.Contains(strFile, fileStr) {
			log.Debug("path=%s is ignore handle err, mode=%s", strFile, fileStr)
			return true
		}

		oneRegex, ok := g.IgnoreErrorFileOrFloderRegexp[fileStr]
		if !ok {
			continue
		}

		if oneRegex.MatchString(strFile) {
			log.Debug("path=%s is ignore handle floder err, mode=%s", fileStr, strFile)
			return true
		}
		// if ok, _ := regexp.Match(fileStr, []byte(strFile)); ok {
		// 	log.Debug("path=%s is ignore handle file err, mode=%s", fileStr, strFile)
		// 	return true
		// }
	}

	// 忽略指定文件中的指定类型错误, 第二层map为忽略的错误类型
	for fileStr, fileTypes := range g.IgnoreFileErrTypesMap {
		if strings.Contains(strFile, fileStr) {
			if _, ok := fileTypes[(int)(errType)]; ok {
				log.Debug("path=%s is ignore special err, errType=%d", strFile, errType)
				return true
			}
		}

		oneRegex, ok := g.IgnoreFileErrTypesRegexp[fileStr]
		if !ok {
			continue
		}

		if oneRegex.MatchString(strFile) {
			_, ok3 := fileTypes[(int)(errType)]
			if ok3 {
				log.Debug("path=%s is ignore special err, errType=%d", strFile, errType)
				return true
			}
		}

		// if ok, _ := regexp.Match(fileStr, []byte(strFile)); ok {
		// 	_, ok3 := fileTypes[(int)(errType)]
		// 	if ok3 {
		// 		log.Debug("path=%s is ignore special err, errType=%d", strFile, errType)
		// 		return true
		// 	}
		// }
	}

	return false
}

// isIgnoreFloder 判断文件夹是否为过滤处理的文件夹
func (g *GlobalConfig) isIgnoreFloder(path string) bool {
	for _, floderStr := range g.IgnoreHandleFolderVec {
		if strings.Contains(path, floderStr) {
			log.Debug("path=%s is ignore handle floder, mode=%s", path, floderStr)
			return true
		}

		if ok, _ := regexp.Match(floderStr, []byte(path)); ok {
			log.Debug("path=%s is ignore handle floder, mode=%s", path, floderStr)
			return true
		}
	}

	return false
}

// isIgnoreFile 判断lua文件是否为过滤处理的文件
func (g *GlobalConfig) isIgnoreFile(path string) bool {
	for _, fileStr := range g.IgnoreHandleFileVec {
		if strings.Contains(path, fileStr) {
			log.Debug("path=%s is ignore handle file, mode=%s", path, fileStr)
			return true
		}

		if ok, _ := regexp.Match(fileStr, []byte(path)); ok {
			log.Debug("path=%s is ignore handle file, mode=%s", path, fileStr)
			return true
		}
	}

	return false
}

// IsHandleAsLua 判断给定的文件的完整文件名，判断是否关联成lua文件，是否要进行分析
func (g *GlobalConfig) IsHandleAsLua(strFile string) bool {
	// 判断直接以.lua为后缀的文g.GetDirManager().GetMainDir()件
	if ok := strings.HasSuffix(strFile, ".lua"); ok {
		return true
	}

	// 判断是否为其他的后缀关联到了lua类型
	if associalStr := g.GetAssocialLua(strFile); associalStr != "" {
		return true
	}

	return false
}

// IsIgnoreCompleteFile 给定完整的文件路径，以及lua项目的根目录，判断是否需要屏蔽分析该文件
func (g *GlobalConfig) IsIgnoreCompleteFile(strFile string) bool {
	dirManager := g.dirManager
	// 如果主dir没有，忽略
	mainDir := dirManager.GetMainDir()
	if mainDir == "" {
		return false
	}

	// 只有主的dir包含在文件里面， 才进行判断，因为主的才会配置忽略
	if !strings.HasPrefix(strFile, mainDir) {
		return false
	}

	// 去掉前缀的文件名
	strTrim := strings.TrimPrefix(strFile, mainDir)
	if g.isIgnoreFile(strTrim) {
		log.Debug("strFile is ignore handle file=%s", strTrim)
		return true
	}

	if g.isIgnoreFloder(strTrim) {
		log.Debug("strFile is ignore handle floder=%s", strTrim)
		return true
	}

	return false
}

// GetStrProtocol 判断给定的字符串是否有协议组，例如c2s, s2s
// 传入的字符串可能为 !c2s，首先去掉! 前缀, 如果找到了返回去掉！前缀的字符串
func (g *GlobalConfig) GetStrProtocol(str string) string {
	if len(str) < 1 {
		return ""
	}

	if str[0] == '!' {
		str = str[1:]
	}

	for _, strVars := range g.ProtocolVars {
		if str == strVars {
			return str
		}
	}

	return ""
}

// IsStrProtocol 判断给定的字符串是否是协议组中的，例如c2s, s2s
// 传入的字符为c2s或是s2s
func (g *GlobalConfig) IsStrProtocol(str string) bool {
	for _, strVars := range g.ProtocolVars {
		if str == strVars {
			return true
		}
	}

	return false
}

// GetRemoveProtocolStr 判断给定字符的前缀是否有协议组
func (g *GlobalConfig) GetRemoveProtocolStr(str string) (strPre, strSuf string) {
	if len(str) < 1 {
		return
	}

	for _, strVars := range g.ProtocolVars {
		if !strings.HasPrefix(str, strVars+".") {
			continue
		}

		strPre = strVars
		strSuf = str[len(strVars)+1:]
		return
	}

	return
}

// IsProtocolSubString 判断当前字符是否为某个协议前缀的子字符串
func (g *GlobalConfig) IsProtocolSubString(str string) bool {
	if len(str) < 1 {
		return false
	}

	for _, strVars := range g.ProtocolVars {
		if strings.Contains(strVars, str) {
			return true
		}
	}
	return false
}

// ClearCacheFileMap 新的扫描和当文件保存的时候，清除掉缓存
func (g *GlobalConfig) ClearCacheFileMap() {
	g.FileExistCacheMap = map[string]bool{}
}

// FileExistCache cache中查找文件是否存在，如果cache中存在直接返回，避免同一个文件读取判断是否存在
func (g *GlobalConfig) FileExistCache(path string) bool {
	g.FileExistCacheMutex.Lock()
	defer g.FileExistCacheMutex.Unlock()

	// cache中存在，直接返回
	flag, ok := g.FileExistCacheMap[path]
	if ok {
		return flag
	}

	// 查找文件真实是否存在，若存在放入cache中
	flag = filefolder.IsFileExist(path)
	if flag {
		g.FileExistCacheMap[path] = true
	} else {
		g.FileExistCacheMap[path] = false
	}

	return flag
}

// IsIgnoreRequireModuleError 判断require不存在的模块，是否需要告警
func (g *GlobalConfig) IsIgnoreRequireModuleError(strName string) bool {
	return g.IgnoreRequireSystemModule[strName]
}

// RebuildSameFileNameVar 重构应该忽略的同文件名的变量
// allFilesMap为最新的加载的所有文件名，包含了前缀
func (g *GlobalConfig) RebuildSameFileNameVar(allFilesMap map[string]string) {
	if !g.IgnoreFileNameVarFlag {
		return
	}

	g.IgnoreSameFileNameMap = map[string]bool{}

	// 获取后缀 /
	for strFile := range allFilesMap {
		lastIndex := strings.LastIndex(strFile, "/")

		if lastIndex > -1 {
			strFile = strFile[lastIndex+1:]
		}

		strFile = strings.TrimSuffix(strFile, ".lua")
		g.IgnoreSameFileNameMap[strFile] = true
	}
}

// IsIgnoreNameVar 判断是否要忽略找不到指定的定义
func (g *GlobalConfig) IsIgnoreNameVar(strName string) bool {
	if _, ok := g.IgnoreVarMap[strName]; ok {
		return true
	}

	if _, ok := g.IgnoreSameFileNameMap[strName]; ok {
		return true
	}

	for _, wildcarStr := range g.IgnoreWildcarVarMap {
		// 匹配到了这个通配符
		if ok, _ := path.Match(wildcarStr, strName); ok {
			return true
		}
	}

	return false
}

// GetGVarExtendFlag 查询_G.a 这样的全局符号，a是否会扩大到全局符号定义
// 例如前面定义了a=1,  那么此时_G.a 会指向前面的a=1
// 获取标记
func (g *GlobalConfig) GetGVarExtendFlag() bool {
	return g.GVarExtendGlobalFlag
}

// GetJudgeColonFlag 是否区分 ：与 . 成员的调用标记。0表示.可以调用:用法；1表示两者不能相互调用；2表示两者可以相互调用
func (g *GlobalConfig) GetJudgeColonFlag() int {
	return g.colonFlag
}

// IsHasProjectEntryFile 是否有工程入口文件
func (g *GlobalConfig) IsHasProjectEntryFile() bool {
	return len(g.ProjectFiles) == 0
}

// IsFrameReferOtherFile 判断传入的函数是否为新增的框架引入其他文件的方式
func (g *GlobalConfig) IsFrameReferOtherFile(strFile string) bool {
	for _, oneReferFrame := range GConfig.ReferFrameFiles {
		if oneReferFrame.Name == strFile {
			return true
		}
	}

	return false
}

// IsImportSuffixFlag 引起其他的框架文件，判断是否要包括.lua后缀
func (g *GlobalConfig) IsImportSuffixFlag(referNameStr string) bool {
	for _, oneReferFrame := range GConfig.ReferFrameFiles {
		if oneReferFrame.Name != referNameStr {
			continue
		}

		if oneReferFrame.SuffixFlag == 0 {
			return false
		}

		return true
	}

	return true
}

// GetReferFrameSubType 获取自定义引入其他的文件，具体引入的子类型
func (g *GlobalConfig) GetReferFrameSubType(referNameStr string) (referSubType ReferFrameType) {
	referSubType = RtypeImport

	for _, oneReferFrame := range GConfig.ReferFrameFiles {
		if oneReferFrame.Name != referNameStr {
			continue
		}

		if oneReferFrame.Type == 0 {
			referSubType = RtypeImport
		} else if oneReferFrame.Type == 1 {
			referSubType = RtypeRequire
		} else if oneReferFrame.Type == 2 {
			referSubType = RtypeAuto
		}

		break
	}

	return referSubType
}

// GetAllReferFileTypes 获取所有的引入文件方式的列表
func (g *GlobalConfig) GetAllReferFileTypes() (strArray []string) {
	strArray = append(strArray, "dofile")
	strArray = append(strArray, "loadfile")
	strArray = append(strArray, "require")
	for _, oneReferFrame := range GConfig.ReferFrameFiles {
		strArray = append(strArray, oneReferFrame.Name)
	}

	return strArray
}

// GetFrameReferFiles 获取框架中引入的其他文件方式的列表
func (g *GlobalConfig) GetFrameReferFiles() (strArray []string) {
	for _, oneReferFrame := range GConfig.ReferFrameFiles {
		strArray = append(strArray, oneReferFrame.Name)
	}

	return strArray
}

// GetPathSeparator 获取路径分割符
func (g *GlobalConfig) GetPathSeparator() string {
	return g.PathSeparator
}

// IsGlobalIgnoreErrType 判断指定类型的告警是否被全局屏蔽了
func (g *GlobalConfig) IsGlobalIgnoreErrType(errorType CheckErrorType) bool {
	_, flag := g.IgnoreErrorTypeMap[errorType]
	return flag
}

// SetAssocialList 设置所有后缀文件关联到lua类型, 例如有的.txt后缀文件，会当成lua文件处理
func (g *GlobalConfig) SetAssocialList(associalList []string) {
	g.AssocialList = associalList
}

// GetAssocialList 获取所有的关联的list
func (g *GlobalConfig) GetAssocialList() (associalList []string) {
	return g.AssocialList
}

// GetAssocialLua 判断文件名是否为关联的lua文件
func (g *GlobalConfig) GetAssocialLua(fileName string) (associalStr string) {
	// 处理下不同平台的路径转换
	filePathStr, _ := filepath.Abs(fileName)
	for _, str := range g.AssocialList {
		if ok, _ := path.Match(str, filePathStr); ok {
			return str
		}
	}
	return
}

// IsIgnoreLocNotUseVar 判断字符串是否为全局模块忽略的局部变量定义了未使用的变量
func (g *GlobalConfig) IsIgnoreLocNotUseVar(strName string) bool {
	_, flag := g.IgnoreLocalNoUseVarMap[strName]
	return flag
}

// IsIgnoreProtocolPreVar 获取协议前缀变量未找到，是否忽略告警
func (g *GlobalConfig) IsIgnoreProtocolPreVar() bool {
	return g.ProtocolPreIngoreFlag
}

// GetDirManager get dir manager
func (g *GlobalConfig) GetDirManager() *DirManager {
	return g.dirManager
}

// MatchAnnotateSet 匹配配置的注解推导类型
func (g *GlobalConfig) MatchAnnotateSet(funcName string) (flag bool, oneSet AnntotateSet) {
	flag = false

	for _, oneValue := range g.anntotateSets {
		if oneValue.FuncName == funcName {
			return true, oneValue
		}
	}

	return
}

// IsSpecialCheck 判断是否需要进行特殊的关联检查
func (g *GlobalConfig) IsSpecialCheck() bool {
	if !g.showWarnFlag {
		return false
	}

	// 需要特殊校验的告警类型
	errTypeList := []CheckErrorType{CheckErrorNoDefine, CheckErrorCycleDefine, CheckErrorCallParam,
		CheckErrorImportVar, CheckErrorNotIfVar}

	// 判断是否屏蔽了上面的告警类型
	for _, errType := range errTypeList {
		if _, ok := g.IgnoreErrorTypeMap[errType]; !ok {
			return true
		}
	}

	return false
}

func (g *GlobalConfig) IsInSysNotUseMap(strName string) bool {
	_, ok := g.ignoreSysNoUseMap[strName]
	return ok
}

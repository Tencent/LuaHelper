package common

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// LuaType 变量定义的类型
type LuaType uint8

const (
	_ LuaType = iota
	// LuaTypeNil nil
	LuaTypeNil

	// LuaTypeBool bool
	LuaTypeBool

	// LuaTypeNumber int64, float64
	LuaTypeNumber

	// LuaTypeInter int64
	LuaTypeInter

	// LuaTypeFloat float64
	LuaTypeFloat

	// LuaTypeString string
	LuaTypeString

	// LuaTypeTable table
	LuaTypeTable

	// LuaTypeFunc 函数
	LuaTypeFunc

	// LuaTypeRefer 引用其他的
	LuaTypeRefer

	// LuaTypeAll 什么都有可能是
	LuaTypeAll
)

// ReferExpFlag 值lua变量关联的exp标记，// 默认值为0，指向的ReferExp是否为empty，例如定义的时候 a = nil，值为1, 当被赋值后，值为2
type ReferExpFlag uint8

const (
	// ReferExpFlagDefault 初始值为0
	ReferExpFlagDefault ReferExpFlag = 0

	// ReferExpFlagEmpty 当lua变量的ReferExp定义为nil时候，此时标记为1
	ReferExpFlagEmpty ReferExpFlag = 1

	// ReferExpFlagAsign 当lua变量的ReferExp之前定义为nil，这是又进行了赋值，值为2
	ReferExpFlagAsign ReferExpFlag = 2
)

// TableKeyStruct table构造时候，一个key值的结构，
type TableKeyStruct struct {
	StrKey    string         // key值的字符串名称
	Loc       lexer.Location // key值的位置信息
	StrDetail string         // 提取到的类型值
}

// ExtraSystem 支持其他插件，额外的扩展信息。当变量为系统库的函数例如ipairs 还有一些模块的时候，指向这个数据
type ExtraSystem struct {
	SysNoticeInfo *SystemNoticeInfo // 如果是指向系统的函数，指针指向对应的
	SysModuleInfo *OneModuleInfo    // 如果是执行系统的模块，指针指向对应的
	SysModuleVar  *SystemModuleVar  // 如果是指向系统的变量
}

// ExtraGlobal 当变量指向的是全局变量时候，扩展信息
type ExtraGlobal struct {
	Prev        *VarInfo     // 当用map管理的时候，指向的前一个全局信息，因为全局信息扫描时候，可能先扫描的不是最初的定义
	ExtraSystem *ExtraSystem // 其他插件，额外的扩展信息。如果这个全局变量为系统的，或是模块的，指向额外的扩展信息
	StrProPre   string       // 表示为协议的前缀 为c2s 或是s2s，项目中定制的
	FuncLv      int          // 函数的层级，主函数层级为0，子函数+1
	ScopeLv     int          // 所在的scop层数，相对于自己所处的func而言
	GFlag       bool         // 是否为_G 类型的变量
}

// ForCycleInfo 变量如果是由for函数引起的，关联的for表达式信息
// for k,v in pairs(b) do block end
// 其中变量k或v会尝试关联下面信息
type ForCycleInfo struct {
	Exp        ast.Exp // 关联的pair或pairs里面的表达式
	IpairsFlag bool    // 是否为ipairs标记
}

// VarInfo 变量信息
type VarInfo struct {
	FileName        string              // 出现的lua文件
	ReferExp        ast.Exp             // 当变量为引用类型的时候，指向引用的指针
	ReferInfo       *ReferInfo          // 指向是否引用了其他的lua文件, nil表示没有指向
	ReferFunc       *FuncInfo           // 如果变量是定义的函数类型，指向函数定义的func，nil表示没有指向
	SubMaps         map[string]*VarInfo // 包含的所有成员信息,构造的，例如 local a = {} a.b = 1, 其中b就是a的成员信息，绑定起来
	ExpandStrMap    map[string]struct{} // 字符串展开存储的字符串map。// local a = { }// print(a.b.c) // a变量的身上挂了一个字符串属性：b.c; ss.data , 当输入ss.时候，自动代码提示data
	ExtraGlobal     *ExtraGlobal        // 当变量指向全局变量时候，扩展的全局变量信息，如果为nil，表示局部变量。
	Loc             lexer.Location      // 初始定义的位置信息
	NoUseAssignLocs []lexer.Location    // 局部变量定义了，未直接使用，后面有对其赋值，记录下所有的位置
	ForCycle        *ForCycleInfo       // 关联的for循环的表达式
	VarType         LuaType             // 变量定义的类型
	VarIndex        uint8               // 当一行语句声明了多个变量时候，例如 local a, b 语句，显示变量的index，默认的为1，例子中a的index为1，b的index为2
	IsParam         bool                // 是否为函数定义的参数，默认为false
	IsForParam      bool                // 是否为泛型for语句产生的参数，默认为false
	IsUse           bool                // 局部变量是否被引用使用过了,第一轮检测的时候，若没有使用进行告警
	IsExpEmpty      bool                // 默认为false，指向的ReferExp是否为empty，例如定义的时候 a = nil， 那么IsExpEmpty为true, 当被赋值后，就不为true
	IsMemFlag       bool                // 是否为其他的变量的成员变量，默认为false
	IsClose         bool                // 是否为lua5.4 close熟悉的变量
}

// VarGetFlag 变量信息获取的方式
type VarGetFlag uint8

const (
	_ VarGetFlag = iota

	// FirstVarFlag 表示先获取的变量类型
	FirstVarFlag

	// FirstAnnotateFlag 表示先获取的注解类型
	FirstAnnotateFlag
)

// Symbol 符号统一的结构，包含了变量出现的文件名、注解类型
type Symbol struct {
	FileName        string           // lua所在的文件名
	VarInfo         *VarInfo         // 变量信息
	AnnotateType    annotateast.Type // 这个变量关联的注解类型
	VarFlag         VarGetFlag       // 变量获取的方式
	AnnotateLine    int              // 如果是注解类型，附上注解类型的行号
	AnnotateLoc     lexer.Location   // 如果是注解类型，获取注解类型的位置信息
	AnnotateComment string           // 注解引入的注释说明
	StrPreComment   string           // 注解引入的hover或complete前置内容
	StrPreClassName string           // 是注解class里面的field成员，值为class的name
}

// GetDefaultSymbol 获取默认的symbol指针
func GetDefaultSymbol(fileName string, varInfo *VarInfo) *Symbol {
	symbol := &Symbol{
		FileName:     fileName,
		VarInfo:      varInfo,
		AnnotateType: nil,
		VarFlag:      FirstVarFlag,
		AnnotateLine: 0,
	}

	return symbol
}

// GetLine 获得统一变量结构出现的行号
func (v *Symbol) GetLine() int {
	if v.VarFlag == FirstVarFlag {
		return v.AnnotateLine
	}

	if v.VarInfo != nil {
		// 如果是变量，注解的行号在变量的前一行
		return v.VarInfo.Loc.StartLine - 1
	}

	return v.AnnotateLine
}

// FindExpFile 根据变量引用，查找到的一个信息
// 引用到的信息，防止递归直接或间接引用自己
type FindExpFile struct {
	FileName string  // lua所在的文件名
	FindExp  ast.Exp // 变量信息
}

// VarInfoList 列表存放
type VarInfoList struct {
	VarVec []*VarInfo // 整体引用的vector，按先后顺序插入到其中
}

// GetLastOneVar 查找最后一个OneGlobalInfo
func (symList *VarInfoList) GetLastOneVar() *VarInfo {
	listLen := len(symList.VarVec)
	if listLen == 0 {
		return nil
	}

	return symList.VarVec[listLen-1]
}

// CreateVarInfo 创建一个VarInfo 指针
// varIndex 表示申明变量的index， 例如local a, b 其中a的index为1，b的index为2。默认为1
func CreateVarInfo(fileName string, varType LuaType, exp ast.Exp, loc lexer.Location, varIndex uint8) *VarInfo {
	// 这里尽量用uint8类型来存，降低内容占用
	// 极端情况下，index可能会超过255个, 转换之后至少为1
	if varIndex == 0 {
		varIndex = 1
	}

	locVar := &VarInfo{
		FileName:   fileName,
		ReferExp:   exp,
		ReferInfo:  nil,
		ReferFunc:  nil,
		SubMaps:    nil,
		Loc:        loc,
		VarIndex:   varIndex,
		VarType:    varType,
		IsParam:    false,
		IsForParam: false,
		IsUse:      false,
		IsExpEmpty: false,
	}

	return locVar
}

// CreateOneVarInfo 用于之前的GlobalInfo
// varIndex 表示申明变量的index， 例如local a, b 其中a的index为1，b的index为2。默认为1
func CreateOneVarInfo(fileName string, loc lexer.Location, referInfo *ReferInfo, referFunc *FuncInfo, varIndex uint8) *VarInfo {
	// 这里尽量用uint8类型来存，降低内容占用
	// 极端情况下，index可能会超过255个, 转换之后至少为1
	if varIndex == 0 {
		varIndex = 1
	}

	locVar := &VarInfo{
		FileName:   fileName,
		ReferInfo:  referInfo,
		ReferFunc:  referFunc,
		SubMaps:    nil,
		Loc:        loc,
		VarIndex:   varIndex,
		VarType:    LuaTypeRefer,
		IsParam:    false,
		IsUse:      false,
		IsForParam: false,
		IsExpEmpty: false,
	}

	return locVar
}

// CreateOneGlobal 生成一个新的GlobalInfo
func CreateOneGlobal(fileName string, funcLv int, scopeLv int, loc lexer.Location, gFlag bool,
	referInfo *ReferInfo, referFunc *FuncInfo, luaFileStr string) *VarInfo {

	locVar := &VarInfo{
		FileName:   fileName,
		ReferInfo:  referInfo,
		ReferFunc:  referFunc,
		SubMaps:    nil,
		Loc:        loc,
		VarIndex:   1,
		VarType:    LuaTypeRefer,
		IsParam:    false,
		IsUse:      false,
		IsForParam: false,
		IsExpEmpty: false,
	}

	extraGlobal := &ExtraGlobal{
		Prev:      nil,
		StrProPre: "",      // 表示是否为协议的前缀 为c2s 或是s2s
		FuncLv:    funcLv,  // 函数的层级，主函数层级为0，子函数+1
		ScopeLv:   scopeLv, // 所在的scop层数，相对于自己所处的func而言
		GFlag:     gFlag,
	}

	locVar.ExtraGlobal = extraGlobal
	return locVar
}

// GetVarTypeDetail 获取类型描述信息
func (varInfo *VarInfo) GetVarTypeDetail() string {
	return GetLuaTypeString(varInfo.VarType, varInfo.ReferExp)
}

// IsHasTabkeKeyStr 查看LocVarInfo是否为指向Table，如果是指向Table，判断是否有传人的key值字符串
func (varInfo *VarInfo) IsHasTabkeKeyStr(strTableKey string) (bool, lexer.Location) {
	var exp ast.Exp
	if varInfo.VarType == LuaTypeTable {
		exp = varInfo.ReferExp
	} else {
		return false, lexer.Location{}
	}

	tableVec := GetTableExpKeyStrVec(exp)
	for _, oneKey := range tableVec {
		if oneKey.StrKey == strTableKey {
			return true, oneKey.Loc
		}
	}

	return false, lexer.Location{}
}

// GetTableKeyTypeDetail 获取一个key的类型信息
func (varInfo *VarInfo) GetTableKeyTypeDetail(strTableKey string) string {
	var exp ast.Exp
	if varInfo.VarType == LuaTypeTable {
		exp = varInfo.ReferExp
	} else {
		return ""
	}

	tableVec := GetTableExpKeyStrVec(exp)
	for _, oneKey := range tableVec {
		if oneKey.StrKey == strTableKey {
			return oneKey.StrDetail
		}
	}

	return ""
}

// IsExistMember 判断局部对象是否含有特定的成员
func (varInfo *VarInfo) IsExistMember(strName string) bool {
	if varInfo.SubMaps == nil {
		return false
	}

	_, ok := varInfo.SubMaps[strName]
	return ok
}

// InsertSubMember 局部里面构建一个成员信息
func (varInfo *VarInfo) InsertSubMember(strName string, subSymbol *VarInfo) {
	if varInfo.SubMaps == nil {
		varInfo.SubMaps = map[string]*VarInfo{}
	}

	subSymbol.IsMemFlag = true
	varInfo.SubMaps[strName] = subSymbol
}

// IsGlobal 判断是否为全局变量
func (varInfo *VarInfo) IsGlobal() bool {
	return varInfo.ExtraGlobal != nil
}

// IsGFlag 判断是否是为_G的全局变量
func (varInfo *VarInfo) IsGFlag() bool {
	if varInfo.ExtraGlobal == nil {
		return false
	}

	return varInfo.ExtraGlobal.GFlag
}

// MakeVarMemTable 当变量为table，构造所有的全局成员变量
// 定义了全局变量 a
// 1)例如a = { bc = 1, cd = 2}
// 2)例如a = a or { bc = 1, cd = 2}
func (varInfo *VarInfo) MakeVarMemTable(valExp ast.Exp, fileName string, loc lexer.Location) {
	tableNode := GetMakeTableConstructorExp(valExp)
	if tableNode == nil {
		return
	}

	for i, keyExp := range tableNode.KeyExps {
		if keyExp == nil {
			continue
		}

		valExp := tableNode.ValExps[i]
		varKeyStr := GetExpName(keyExp)
		if !JudgeSimpleStr(varKeyStr) {
			continue
		}

		if varInfo.IsExistMember(varKeyStr) {
			continue
		}

		loc := GetExpLoc(keyExp)
		newVar := CreateOneVarInfo(fileName, loc, nil, nil, 1)
		newVar.VarType = GetExpType(valExp)
		newVar.ReferExp = valExp
		varInfo.InsertSubMember(varKeyStr, newVar)
	}
}

// HasSubVarInfo 判断是否有子是subVar
func (varInfo *VarInfo) HasSubVarInfo(subSymbol *VarInfo) bool {
	if varInfo == subSymbol {
		return true
	}

	if varInfo.SubMaps != nil {
		for _, oneVarInfo := range varInfo.SubMaps {
			if oneVarInfo == subSymbol {
				return true
			}
		}
	}

	return false
}

// FindAllVar 找到当前varinfo 下的所有 local function | variable
func (varInfo *VarInfo) FindAllVar(strName string, preStrName string) (symbolStruct FileSymbolStruct) {
	if preStrName == "" {
		return
	}

	fullName := preStrName + "." + strName
	if varInfo.ReferFunc != nil && varInfo.ReferFunc.IsColon {
		fullName = preStrName + ":" + strName
	}

	symbolStruct = FileSymbolStruct{
		Name: fullName,
		Kind: IKVariable,
		Loc:  varInfo.Loc,
	}

	if varInfo.ReferFunc != nil {
		symbolStruct.Kind = IKFunction
		symbolStruct.Loc = varInfo.ReferFunc.Loc
		paramStr := ""
		// symbolStruct.Children = varInfo.ReferFunc.MainScope.FindAllLocalVal(nil)
		for _, param := range varInfo.ReferFunc.ParamList {
			if param == "self" {
				continue
			}
			if paramStr == "" {
				paramStr += param
			} else {
				paramStr += ", " + param
			}
		}
		if paramStr != "" {
			symbolStruct.Name = symbolStruct.Name + "(" + paramStr + ")"
		}
	}

	return
}

// IsCorrectPosition 判断当前varinfo.loc 是否在查找的loc 之前
func (varInfo *VarInfo) IsCorrectPosition(loc lexer.Location) bool {
	// 当前varinfo.Loc 不在loc 之前 直接返回false
	if !varInfo.Loc.IsBeforeLoc(loc) {
		return false
	}

	switch varInfo.ReferExp.(type) {
	// 若是*ast.FuncDefExp 、 *ast.NameExp、 *ast.NameExp 需要做额外的检测
	// 给出一个范围，判断这个范围是否包含在这表达式中
	// 1. local bbbccc = bbbccc
	// 2.
	// local bbbccc = function()
	// 	bbbccc()
	// end
	case *ast.FuncDefExp:
		exp := varInfo.ReferExp.(*ast.FuncDefExp)
		// 若是直接定义的function， 则直接返回true
		// local function bbbccc()
		// 	bbbccc()
		// end
		if exp.Loc.IsContainLoc(varInfo.Loc) {
			return true
		}

		if exp.Loc.IsContainLoc(loc) {
			return false
		}
		return true
	case *ast.NameExp:
		exp := varInfo.ReferExp.(*ast.NameExp)
		if exp.Loc.IsContainLoc(loc) {
			return false
		}
		return true
	case *ast.FuncCallExp:
		exp := varInfo.ReferExp.(*ast.FuncCallExp)
		if exp.Loc.IsContainLoc(loc) {
			return false
		}
		return true
	default:
		// 其余类型，直接返回true
		return true
	}
}

// HasDeepSubVarInfo 深入判断子成员是否包含另外一个变量
func (varInfo *VarInfo) HasDeepSubVarInfo(subSymbol *VarInfo) bool {
	for _, oneVar := range varInfo.SubMaps {
		if oneVar == subSymbol {
			return true
		}

		if oneVar.HasDeepSubVarInfo(subSymbol) {
			return true
		}
	}

	return false
}

func findValidKey(oneExp *ast.TableConstructorExp, strTableKey string, line int, charactor int) bool {
	for index, keyExp := range oneExp.KeyExps {
		if keyExp == nil {
			// 查看value
			if len(oneExp.ValExps) <= index {
				continue
			}

			valueExp := oneExp.ValExps[index]
			nameExp, ok := valueExp.(*ast.NameExp)
			if !ok {
				continue
			}

			if nameExp.Name != strTableKey {
				continue
			}

			if nameExp.Loc.IsInLocStruct(line, charactor) {
				return true
			}
		} else {
			strExp, ok := keyExp.(*ast.StringExp)
			if !ok {
				continue
			}

			if strExp.Str != strTableKey {
				continue
			}

			if strExp.Loc.IsInLocStruct(line, charactor) {
				return true
			}
		}
	}

	return false
}

// IsHasReferTableKey 判断这个变量是否指向一个table，且它的key是一个strTableKey；或是是key为nil，value为strTableKey
func (varInfo *VarInfo) IsHasReferTableKey(strTableKey string, line int, charactor int) bool {
	referExp := varInfo.ReferExp
	if referExp == nil {
		return false
	}

	oneExp, ok := referExp.(*ast.TableConstructorExp)
	if !ok {
		return false
	}

	if !oneExp.Loc.IsInLocStruct(line, charactor) {
		return false
	}

	if findValidKey(oneExp, strTableKey, line, charactor) {
		return true
	}

	return false
}

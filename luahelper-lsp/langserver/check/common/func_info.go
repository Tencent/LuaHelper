package common

import (
	//"bytes"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// FuncRelateVar 如果func是冒号的语法，例如
// a = {}
// function a:func1()
// 那么func1需要反向指向的变量（局部变量或是全局的)
type FuncRelateVar struct {
	StrName string // 关联的变量名字
}

// CreateFuncRelateVar 创建函数的反向关联指针
func CreateFuncRelateVar(strName string, relateVarInfo *VarInfo) *FuncRelateVar {
	return &FuncRelateVar{
		StrName: strName,
	}
}

// LabelInfo 定义label标签，lua goto语法用到
type LabelInfo struct {
	name    string         // 名称
	scopeLv int            // 当前的作用域层级，初始值为0
	loc     lexer.Location // 位置信息
}

// ReturnItem 一个返回数据，有可能是返回整数或是其他的数值，那么LocVar和GlobalVar都为nil
type ReturnItem struct {
	VarType   LuaType  // 返回值的类型
	VarInfo   *VarInfo // 关联的变量指针
	ReturnExp ast.Exp  // 关联的exp
}

// ReturnInfo 函数返回信息
type ReturnInfo struct {
	ReturnVarVec []ReturnItem // 函数一次返回可能返回多个字段，这里有列表存储
}

// FuncInfo 函数信息
type FuncInfo struct {
	parent           *FuncInfo           // 父的funcInfo
	labelVecs        []*LabelInfo        // 包含所有的labelInfo
	ReturnVecs       []*ReturnInfo       // 包含所有的函数返回信息, 函数可能有多处返回，用列表存储
	MainScope        *ScopeInfo          // 函数指向的主的ScopeInfo
	RelateVar        *FuncRelateVar      // 函数反向关联的指针，当为冒号函数时候，即IsColon为true才存储
	ParamList        []string            // 函数所有的参数列表
	Loc              lexer.Location      // 位置信息
	ScopeLv          int                 // 当前的作用域层级，初始值为0
	FuncLv           int                 // func的层级，最上层的func层级为0，子的func层级+1
	FuncID           int                 // funcInfo在AnalysisFileResult中出现的序号，默认从0开始
	IsVararg         bool                // 是否含义可变参数
	IsColon          bool                // 是否为: 这样的函数
	ParamDefaultNum  int                 // 默认函数数量
	ParamDefaultInit bool                // 是否获取过默认函数数量
	FileName         string              // 函数所在的文件名
	ClassName        string              // 例如 function table.func() end // table即ClassName
	FuncName         string              // 例如 function table.func() end // func即FuncName
	ParamType        map[string][]string // 函数所有的参数注解类型列表 参数可能有多个类型 number|string
	ReturnType       [][]string          // 函数注解处的返回值类型 返回值只能按顺序查找
}

// CreateFuncInfo 创建一个函数指针
func CreateFuncInfo(parent *FuncInfo, funcLv int, loc lexer.Location, isVararg bool, parentScope *ScopeInfo,
	fileName string) *FuncInfo {
	funcInfo := &FuncInfo{
		parent:           parent,
		labelVecs:        nil,
		MainScope:        nil,
		RelateVar:        nil,
		ParamList:        nil,
		Loc:              loc,
		ScopeLv:          0,
		FuncLv:           funcLv,
		FuncID:           0,
		IsVararg:         isVararg,
		IsColon:          false,
		ParamDefaultNum:  -1,
		ParamDefaultInit: false,
		FileName:         fileName,
		ParamType:        make(map[string][]string),
	}

	// 设置函数指向的主scope
	mainScope := CreateScopeInfo(parentScope, funcInfo, loc)
	funcInfo.MainScope = mainScope

	if parentScope != nil {
		parentScope.AppendSubScope(mainScope)
	}

	return funcInfo
}

// EnterScope scope层+1
func (fun *FuncInfo) EnterScope() {
	fun.ScopeLv++
}

// ExitScope scope层-1
func (fun *FuncInfo) ExitScope() {
	fun.ScopeLv--
}

// InsertLabel 在labelVecs插入定义的标签
func (fun *FuncInfo) InsertLabel(name string, loc lexer.Location) {
	newLabel := &LabelInfo{
		name:    name,
		scopeLv: fun.ScopeLv,
		loc:     loc,
	}

	if fun.labelVecs == nil {
		fun.labelVecs = []*LabelInfo{}
	}

	fun.labelVecs = append(fun.labelVecs, newLabel)
}

// FindLabel 在labelVecs中查找标签
func (fun *FuncInfo) FindLabel(name string, curScopeLv int) bool {
	for _, label := range fun.labelVecs {
		if label.name != name {
			continue
		}

		if label.scopeLv <= curScopeLv {
			return true
		}
	}

	return false
}

// InsertReturn 向函数返回值列表中插入一种返回值的情况
func (fun *FuncInfo) InsertReturn(returnInfo *ReturnInfo) {
	fun.ReturnVecs = append(fun.ReturnVecs, returnInfo)
}

// GetOneReturnVar 获取函数的第一个返回值
func (fun *FuncInfo) GetOneReturnVar() (varInfo *VarInfo) {
	if len(fun.ReturnVecs) < 1 {
		return
	}

	returnInfo := fun.ReturnVecs[0]
	if len(returnInfo.ReturnVarVec) < 1 {
		return
	}

	oneReturn := returnInfo.ReturnVarVec[0]
	varInfo = oneReturn.VarInfo
	return
}

// GetOneReturnExp 获取函数返回的第一个Exp
func (fun *FuncInfo) GetOneReturnExp() (flag bool, exp ast.Exp) {
	return fun.GetReturnIndexExp(1)
}

// GetLastReturnExp 获取函数最后一个返回的Exp, require一个文件时候，可能提前返回不合法的
func (fun *FuncInfo) GetLastOneReturnExp() (flag bool, exp ast.Exp) {
	if len(fun.ReturnVecs) < 1 {
		return
	}

	for i := len(fun.ReturnVecs) - 1; i >= 0; i-- {
		oneReturn := fun.ReturnVecs[i]
		if len(oneReturn.ReturnVarVec) >= 1 {
			exp = oneReturn.ReturnVarVec[0].ReturnExp
			// 如果为nil忽略
			if _, ok := exp.(*ast.NilExp); ok {
				continue
			}

			flag = true
			break
		}
	}

	return
}

// GetReturnIndexExp 获取函数指定位置的返回值
// index 默认从1开始，如果为1表示获取第一个返回值
func (fun *FuncInfo) GetReturnIndexExp(index uint8) (flag bool, exp ast.Exp) {
	if len(fun.ReturnVecs) < 1 {
		return
	}

	for _, oneReturn := range fun.ReturnVecs {
		if len(oneReturn.ReturnVarVec) >= (int)(index) {
			exp = oneReturn.ReturnVarVec[index-1].ReturnExp
			// 如果为nil忽略
			if _, ok := exp.(*ast.NilExp); ok {
				continue
			}

			flag = true
			break
		}
	}

	return
}

// GetFuncCompleteStr 获取函数的代码提示，包含函数的参数
// paramTipFlag 表示是否提示函数的参数
// colonFlag 如果是冒号语法，有时候需要忽略掉self
func (fun *FuncInfo) GetFuncCompleteStr(funcName string, paramTipFlag bool, colonFlag bool) string {
	if !paramTipFlag {
		return funcName
	}

	funcName += "("
	preFlag := false
	for index, oneParam := range fun.ParamList {
		if colonFlag && fun.IsColon && index == 0 {
			continue
		}

		if preFlag {
			funcName += ", "
		}
		funcName += oneParam
		preFlag = true
	}

	if fun.IsVararg {
		if preFlag {
			funcName += ", "
		}
		funcName += "..."
	}

	funcName += ")"
	return funcName
}

// GetFuncCompleteStr 获取函数的代码提示，包含函数的参数
// paramTipFlag 表示是否提示函数的参数
func (fun *FuncInfo) GetFuncCompleteStrInclude(funcName string, paramTipFlag bool) string {
	if !paramTipFlag {
		return funcName
	}

	funcName += "( self"
	preFlag := false
	for _, oneParam := range fun.ParamList {
		funcName += ", "
		funcName += oneParam
		preFlag = true
	}

	if fun.IsVararg {
		if preFlag {
			funcName += ", "
		}
		funcName += "..."
	}

	funcName += ")"
	return funcName
}

// FindFirstColonFunc 向上找到第一个含义冒号语法的函数
func (fun *FuncInfo) FindFirstColonFunc() *FuncInfo {
	if fun.IsColon {
		return fun
	}

	if fun.parent == nil {
		return nil
	}

	// 向上找父函数
	return fun.parent.FindFirstColonFunc()
}

func (fun *FuncInfo) GetParent() *FuncInfo {
	return fun.parent
}

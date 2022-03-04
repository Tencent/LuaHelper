package common

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// NotValStruct 存放这个块中所有的 if not a
type NotValStruct struct {
	Var     *VarInfo // 指向的局部的变量
	SetFlag bool     // 是否已经赋值过了
}

// ScopeInfo 作用域，块信息
type ScopeInfo struct {
	Parent       *ScopeInfo              // 父的ScopeInfo
	SubScopes    []*ScopeInfo            // 所有子的ScopeInfos
	Func         *FuncInfo               // 是否对应的函数信息
	LocVarMap    map[string]*VarInfoList // 所有的局部信息
	NotVarMap    map[string]NotValStruct // 存放这个块中所有的 if not a ; a这样的局部变量，检查后面的语法
	MidNotVarMap map[string]NotValStruct // 中间存放的，例如某一个模块里面 if not a or b then，后面又有if not a then，在这个块里面存放临时的数据
	Loc          lexer.Location          // 位置信息
}

// CreateScopeInfo 创建一个ScopeInfo指针
func CreateScopeInfo(parent *ScopeInfo, funcInfo *FuncInfo, loc lexer.Location) *ScopeInfo {
	return &ScopeInfo{
		Parent:       parent,
		Func:         funcInfo,
		Loc:          loc,
		LocVarMap:    nil,
		NotVarMap:    nil,
		MidNotVarMap: nil,
	}
}

// SetMidNotValStruct 设置一个midNotVar值
func (scope *ScopeInfo) SetMidNotValStruct(strName string, notValStruct NotValStruct) {
	if scope.MidNotVarMap == nil {
		scope.MidNotVarMap = map[string]NotValStruct{}
	}

	scope.MidNotVarMap[strName] = notValStruct
}

// SetNotVarMapStruct 设置一个notVar值
func (scope *ScopeInfo) SetNotVarMapStruct(strName string, notValStruct NotValStruct) {
	if scope.NotVarMap == nil {
		scope.NotVarMap = map[string]NotValStruct{}
	}

	scope.NotVarMap[strName] = notValStruct
}

// AppendSubScope 插入一个子的scopeinfo
func (scope *ScopeInfo) AppendSubScope(subScope *ScopeInfo) {
	scope.SubScopes = append(scope.SubScopes, subScope)
}

// AddLocVar 块里面增加一个变量（局部变量）
func (scope *ScopeInfo) AddLocVar(fileName string, name string, LuaType LuaType, exp ast.Exp, loc lexer.Location,
	varIndex uint8) *VarInfo {
	newVar := &VarInfo{
		FileName:  fileName,
		VarType:   LuaType,
		ReferExp:  exp,
		SubMaps:   nil,
		Loc:       loc,
		VarIndex:  varIndex,
		ReferInfo: nil,
		IsParam:   false,
		IsUse:     false,
	}

	scope.InsertLocalVar(name, newVar)
	return newVar
}

// InsertLocalVar scope中插入一个局部变量
func (scope *ScopeInfo) InsertLocalVar(strName string, locVar *VarInfo) {
	if scope.LocVarMap == nil {
		scope.LocVarMap = map[string]*VarInfoList{}
	}

	// 如果是第二轮工程的check，_G的全局符号放入到工程的结构中
	locInfoList := scope.LocVarMap[strName]
	if locInfoList == nil {
		locInfoList := &VarInfoList{
			VarVec: make([]*VarInfo, 0, 1),
		}
		// 之前这个strName _G的符号不存在，创建
		locInfoList.VarVec = append(locInfoList.VarVec, locVar)
		scope.LocVarMap[strName] = locInfoList
	} else {
		// 之前这个strName _G的符号已经存在，插入到vec中
		locInfoList.VarVec = append(locInfoList.VarVec, locVar)
	}
}

// FindLocVar 新的方式查找局部变量
func (scope *ScopeInfo) FindLocVar(name string, loc lexer.Location) (*VarInfo, bool) {
	locInfoList := scope.LocVarMap[name]
	if locInfoList == nil {
		// 当前没有找到，判断上层是否有找的
		if scope.Parent != nil {
			return scope.Parent.FindLocVar(name, loc)
		}

		return nil, false
	}

	// 需要倒序遍历
	for i := len(locInfoList.VarVec) - 1; i >= 0; i-- {
		locVar := locInfoList.VarVec[i]

		if locVar.IsCorrectPosition(loc) {
			return locVar, true
		}
	}

	// 当前没有找到，判断上层是否有找的
	if scope.Parent != nil {
		return scope.Parent.FindLocVar(name, loc)
	}

	return nil, false
}

// 判断点坐标是否在位置范围内
// line从1开始，column从0开始
func isInLocation(loc *lexer.Location, line, column int) bool {
	// 1) 首先判断行是否在范围内
	if line < loc.StartLine || line > loc.EndLine {
		return false
	}

	// 2) 行与起始的一样，判断列的范围
	if line == loc.StartLine {
		if column < (int)(loc.StartColumn) {
			return false
		}
	}

	// 3) 行与结束的一样，判断列的范围
	if line == loc.EndLine {
		if column > (int)(loc.EndColumn) {
			return false
		}
	}

	return true
}

// FindMinScope 通过行与列找到最小单位的ScopInfo
// line 从1开始
// column 从0开始
func (scope *ScopeInfo) FindMinScope(line, column int) (minScope *ScopeInfo) {
	minScope = nil
	// 判断主的scope位置信息是否满足
	if !isInLocation(&scope.Loc, line, column) {
		return minScope
	}

	minScope = scope
	// 遍历所有的子的scopeInfo
	for _, subScope := range scope.SubScopes {
		// 如果子的位置信息太小，继续找大的
		if subScope.Loc.EndLine < line {
			continue
		}

		// 判断子的位置信息是否满足，若满足，递归向下找
		if isInLocation(&subScope.Loc, line, column) {
			minScope = subScope.FindMinScope(line, column)
			// 一个子的位置信息满足了，其他的子的位置信息就一定不满足，退出
			break
		}

		// 如果子的位置信息不满足, 判断子的位置信息是否已经超过了范围
		if subScope.Loc.StartLine > line {
			break
		}
	}

	return minScope
}

// GetCompleteVar 查找scopeInfo范围代码补全，包括父的scopeInfo
func (scope *ScopeInfo) GetCompleteVar(completeVar *CompleteVarStruct, fileName string, loc lexer.Location, cache *CompleteCache) {
	for strName, locInfoList := range scope.LocVarMap {
		if !IsCompleteNeedShow(strName, completeVar) {
			continue
		}

		if cache.ExistStr(strName) {
			continue
		}

		// 从后面向前找
		for index := len(locInfoList.VarVec) - 1; index >= 0; index-- {
			locVar := locInfoList.VarVec[index]

			if locVar.Loc.StartLine > loc.StartLine {
				continue
			}

			if locVar.Loc.StartLine == loc.StartLine && locVar.Loc.StartColumn > loc.StartColumn {
				continue
			}

			cache.InsertCompleteVar(fileName, strName, locVar)

			// 变量放进来
			break
		}
	}

	// 如果父的scope不为空，递归向上找
	if scope.Parent != nil {
		scope.Parent.GetCompleteVar(completeVar, fileName, loc, cache)
	}
}

// FindMinFunc 查找scopeInfo的最近的funcInfo
func (scope *ScopeInfo) FindMinFunc() *FuncInfo {
	if scope.Func != nil {
		return scope.Func
	}

	if scope.Parent != nil {
		return scope.Parent.FindMinFunc()
	}

	return nil
}

// FindTableKeyReferVarName 在cope中查找table的局部变量关联的局部变量名称
// local a = {
//	 b = 1,
//}
// 如上，当找b的时候关联到变量a
// 返回值表示管理到的局部变量
func (scope *ScopeInfo) FindTableKeyReferVarName(strTableKey string, line int, charactor int) (firstStr, secondStr string) {
	for strName, locInfo := range scope.LocVarMap {
		for _, oneLocInfo := range locInfo.VarVec {
			if oneLocInfo.SubMaps == nil {
				continue
			}

			firstStr, secondStr = GetSubMapStrKey(oneLocInfo.SubMaps, strName, strTableKey, line, charactor)
			if firstStr != "" {
				return firstStr, secondStr
			}
		}
	}

	// 尝试在父的scope中找，如果父的不存在直接返回
	if scope.Parent == nil {
		return "", ""
	}

	return scope.Parent.FindTableKeyReferVarName(strTableKey, line, charactor)
}

// GetTableKeyVar 获取scope下的VarInfo，它是指向一个table，它的key是一个strTableKey；或是是key为nil，value为strTableKey
func (scope *ScopeInfo) GetTableKeyVar(strTableKey string, line int, charactor int) (string, *VarInfo) {
	for strName, locInfo := range scope.LocVarMap {
		for _, oneVar := range locInfo.VarVec {
			if oneVar.IsHasReferTableKey(strTableKey, line, charactor) {
				return strName, oneVar
			}
		}
	}

	// 尝试在父的scope中找，如果父的不存在直接返回
	if scope.Parent == nil {
		return "", nil
	}

	return scope.Parent.GetTableKeyVar(strTableKey, line, charactor)
}

// IsExistLocVarTableStrKey 查找块中所有的局部变量，判断是否有定义的table，包含下面的字符串key
func (scope *ScopeInfo) IsExistLocVarTableStrKey(strTableKey string, line int, charactor int) (firstStr, secondStr string) {
	firstStr, secondStr = scope.FindTableKeyReferVarName(strTableKey, line, charactor)
	if firstStr != "" {
		return firstStr, secondStr
	}

	for strName, locInfo := range scope.LocVarMap {
		for _, oneLocInfo := range locInfo.VarVec {
			ok, findLoc := oneLocInfo.IsHasTabkeKeyStr(strTableKey)
			if !ok {
				continue
			}

			if findLoc.StartLine != line || findLoc.EndLine != line {
				continue
			}

			if findLoc.EndColumn+1 < charactor {
				continue
			}

			if findLoc.StartColumn >= 1 && findLoc.StartColumn-1 > charactor {
				continue
			}

			return strName, ""
		}
	}

	return "", ""
}

// FindNotVarInfo 向上的scope中查找NotVarMap里面的变量，并且返回LocalVarInfo
// checkFlag 表示是否只找SetFlag为false的
func (scope *ScopeInfo) FindNotVarInfo(strName string) (findVar *VarInfo, findScope *ScopeInfo) {
	if scope.NotVarMap == nil {
		return nil, nil
	}

	notValStruct, ok := scope.NotVarMap[strName]
	if !ok {
		// 当前没有找到，判断上层是否有找的
		if scope.Parent != nil {
			return scope.Parent.FindNotVarInfo(strName)
		}

		return nil, nil
	}

	return notValStruct.Var, scope
}

// FindMidNotVarInfo 先只找mid的
func (scope *ScopeInfo) FindMidNotVarInfo(strName string) (findVar *VarInfo, findScope *ScopeInfo) {
	if scope.MidNotVarMap == nil {
		return nil, nil
	}

	notValStruct, ok := scope.MidNotVarMap[strName]
	if !ok {
		// 当前没有找到，判断上层是否有找的
		if scope.Parent != nil {
			return scope.Parent.FindMidNotVarInfo(strName)
		}

		return nil, nil
	}
	return notValStruct.Var, scope
}

// CheckNotVarInfo 只是check
func (scope *ScopeInfo) CheckNotVarInfo(strName string) (findVar *VarInfo, findScope *ScopeInfo) {
	notLocVar, notScope := scope.FindNotVarInfo(strName)
	if notLocVar == nil {
		return nil, nil
	}

	if notScope.NotVarMap == nil {
		return nil, nil
	}

	notValStruct, ok := notScope.NotVarMap[strName]
	if ok {
		if !notValStruct.SetFlag {
			midLocVar, _ := scope.FindMidNotVarInfo(strName)
			if midLocVar != nil {
				return nil, nil
			}

			return notLocVar, notScope
		}
	}

	return nil, nil
}

// GetParamVarInfo 获取参数变量关联的VarInfo
func (scope *ScopeInfo) GetParamVarInfo(paramName string) *VarInfo {
	varInfoList, ok := scope.LocVarMap[paramName]
	if !ok {
		return nil
	}
	if len(varInfoList.VarVec) == 0 {
		return nil
	}

	oneVac := varInfoList.VarVec[0]
	if !oneVac.IsParam {
		return nil
	}

	return oneVac
}

// FindAllLocalVal 获取当前scope下所有local function | variable
func (scope *ScopeInfo) FindAllLocalVal(gScopes []*ScopeInfo) (allSymbolStruct []FileSymbolStruct) {
	// 存储不在当前scope下的subscope 没有在当前层显示的变量与函数
	subScopeLen := len(scope.SubScopes)
	scopeInfos := make(map[*ScopeInfo]bool, subScopeLen)
	for _, scopeInfo := range scope.SubScopes {
		scopeInfos[scopeInfo] = true
	}

	for _, gScope := range gScopes {
		delete(scopeInfos, gScope)
	}

	// 分析所有的local 变量与函数
	for strVar, locVarinfoList := range scope.LocVarMap {
		oneLocInfo := locVarinfoList.VarVec[len(locVarinfoList.VarVec)-1]

		// 获取参数变量
		if oneLocInfo.IsParam {
			continue
		}

		oneSymbol := FileSymbolStruct{
			Name:          strVar,
			Kind:          IKVariable,
			Loc:           oneLocInfo.Loc,
			ContainerName: "local",
		}

		if oneLocInfo.ReferFunc != nil {
			oneSymbol.Kind = IKFunction
			oneSymbol.Loc = oneLocInfo.ReferFunc.Loc
			curChirenScope := oneLocInfo.ReferFunc.MainScope
			paramStr := ""
			for _, param := range oneLocInfo.ReferFunc.ParamList {
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
				oneSymbol.Name = oneSymbol.Name + "(" + paramStr + ")"
			}
			delete(scopeInfos, curChirenScope)
		} else {
			if oneLocInfo.SubMaps == nil {
				allSymbolStruct = append(allSymbolStruct, oneSymbol)
				continue
			}

			var maxLoc lexer.Location
			maxLoc.EndLine = oneSymbol.Loc.EndLine
			maxLoc.EndColumn = oneSymbol.Loc.EndColumn

			// 若当前变量为 表结构，则分析它的子结构
			for strName, varInfo := range oneLocInfo.SubMaps {
				subOneSymbol := varInfo.FindAllVar(strName, strVar)
				if oneSymbol.Children == nil {
					oneSymbol.Children = make([]FileSymbolStruct, 0)
				}
				if subOneSymbol.Name != "" {
					oneSymbol.Children = append(oneSymbol.Children, subOneSymbol)
				}

				if subOneSymbol.Loc.EndLine > maxLoc.EndLine {
					maxLoc.EndLine = subOneSymbol.Loc.EndLine
					maxLoc.EndColumn = subOneSymbol.Loc.EndColumn
				} else if subOneSymbol.Loc.EndLine == maxLoc.EndLine && subOneSymbol.Loc.EndColumn > maxLoc.EndColumn {
					maxLoc.EndColumn = subOneSymbol.Loc.EndColumn
				}

				if varInfo.ReferFunc != nil && varInfo.ReferFunc.MainScope != nil {
					delete(scopeInfos, varInfo.ReferFunc.MainScope)
				}
			}

			oneSymbol.Loc.EndLine = maxLoc.EndLine
			oneSymbol.Loc.StartColumn = maxLoc.EndColumn
		}

		allSymbolStruct = append(allSymbolStruct, oneSymbol)
	}

	// 分析当前scope中中没有分析的 subscope
	for scopeInfo := range scopeInfos {
		symbols := scopeInfo.FindAllLocalVal(nil)
		allSymbolStruct = append(allSymbolStruct, symbols...)
	}

	return
}

// 递归插入子成员
func pushSubMapVars(varInfo *VarInfo, results *[]*VarInfo) {
	if varInfo.SubMaps == nil {
		return
	}

	for _, subVar := range varInfo.SubMaps {
		*results = append(*results, subVar)
		pushSubMapVars(subVar, results)
	}
}

// GetAllVarInfos 递归获取这个ScopeInfo里面的所有变量信息，排查掉函数参数的信息
func (scope *ScopeInfo) GetAllVarInfos(results *[]*VarInfo) {
	// 直接获取个scope内的变量
	for _, symList := range scope.LocVarMap {
		for _, varInfo := range symList.VarVec {
			if varInfo.IsParam {
				continue
			}

			*results = append(*results, varInfo)
			pushSubMapVars(varInfo, results)
		}
	}

	// 递归获取所有子的scope变量
	for _, subScope := range scope.SubScopes {
		subScope.GetAllVarInfos(results)
	}
}

// ForLineVarString 查询scope下的满足的位置信息
func (scope *ScopeInfo) ForLineVarString(line int) (strList []string) {
	for strName, infoList := range scope.LocVarMap {
		for _, oneVar := range infoList.VarVec {
			if !oneVar.IsForParam {
				continue
			}

			if oneVar.Loc.StartLine != line {
				continue
			}

			strList = append(strList, strName)
		}
	}

	for _, subScope := range scope.SubScopes {
		subStrList := subScope.ForLineVarString(line)
		strList = append(strList, subStrList...)
	}

	return strList
}

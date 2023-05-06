package annotateast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// TraverseOneType 遍历解析这个type，获取最简单的type类型字符串，只允许简单
// 如果字符串为空，表示没有匹配到合适的
func TraverseOneType(astType Type) string {
	switch subAst := astType.(type) {
	case *MultiType:
		if len(subAst.TypeList) == 0 {
			return ""
		}

		return TraverseOneType(subAst.TypeList[0])
	case *NormalType:
		return subAst.StrName
	}

	return ""
}

// TypeConvertStr 根据类型获取代码补全时候的提示字符串
func TypeConvertStr(astType Type) string {
	switch subAst := astType.(type) {
	case *MultiType:
		if len(subAst.TypeList) == 0 {
			return ""
		}

		// 有多种类型，是或者的关系, 当有多种可能的类型的，用 | 分割字符串
		multiStr := ""
		for _, oneType := range subAst.TypeList {
			strOne := TypeConvertStr(oneType)
			if strOne == "" {
				continue
			}

			if multiStr != "" {
				multiStr = multiStr + " | "
			}

			multiStr = multiStr + strOne
		}

		return multiStr
	case *NormalType:
		return subAst.StrName

	case *ArrayType:
		return TypeConvertStr(subAst.ItemType) + "[]"

	case *TableType:
		if subAst.EmptyFlag {
			// 空的table
			return "table"
		}

		return "table<" + TypeConvertStr(subAst.KeyType) + ", " + TypeConvertStr(subAst.ValueType) + ">"

	case *FuncType:
		funStr := "function("
		for index, paramStr := range subAst.ParamNameList {
			if index > 0 {
				funStr = funStr + ", "
			}

			paramTypeStr := ""
			if len(subAst.ParamTypeList) > index {
				paramTypeStr = TypeConvertStr(subAst.ParamTypeList[index])
			}

			funStr = funStr + paramStr + ": " + paramTypeStr

		}

		funStr = funStr + ")"

		for index, oneReturn := range subAst.ReturnTypeList {
			if index == 0 {
				funStr = funStr + ": "
			} else {
				funStr = funStr + ", "
			}

			funStr = funStr + TypeConvertStr(oneReturn)
		}
		return funStr
	case *ConstType:
		if subAst.QuotesFlag {
			return "\"" + subAst.Name + "\""
		}
		return subAst.Name
	}

	return ""
}

// FuncTypeConvertStr 函数转换为字符串
// firstParamFlag 是否包含第一个参数的标记。0为默认，不进行更改；1表示去掉第一个参数；2表示第一个参数增加self
func FuncTypeConvertStr(funcType *FuncType, firstParamFlag int) string {
	if funcType == nil {
		return ""
	}

	funStr := "function("
	flag := false
	if firstParamFlag == 2 {
		funStr += "self"
		flag = true
	}
	for index, paramStr := range funcType.ParamNameList {
		if firstParamFlag == 1 && index == 0 {
			continue
		}

		if flag {
			funStr = funStr + ", "
		} else {
			flag = true
		}

		paramTypeStr := ""
		if len(funcType.ParamTypeList) > index {
			paramTypeStr = TypeConvertStr(funcType.ParamTypeList[index])
		}

		funStr = funStr + paramStr + " : " + paramTypeStr

	}

	funStr = funStr + ")"

	for index, oneReturn := range funcType.ReturnTypeList {
		if index == 0 {
			funStr = funStr + " : "
		} else {
			funStr = funStr + ", "
		}

		funStr = funStr + TypeConvertStr(oneReturn)
	}
	return funStr
}

// IsTypeEmpty 判断Type是否为空的, 空的意思是没有进行赋值
func IsTypeEmpty(astType Type) bool {
	if astType == nil {
		return true
	}

	return false
}

// GetAllNormalStrList 获取Type的所有可能的字符串，所有的可能的Type要为NormalType
func GetAllNormalStrList(astType Type) (strList []string) {
	switch subAst := astType.(type) {
	case *MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			tmpList := GetAllNormalStrList(oneType)
			strList = append(strList, tmpList...)
		}

		return strList
	case *NormalType:
		strList = append(strList, subAst.StrName)
		return strList

	case *TableType:
		strList = append(strList, "table")
		return strList
	case *FuncType:
		strList = append(strList, "function")
		return strList
	}

	return
}

// GetAllFuncType 判断是否指向的是一个函数的类型
func GetAllFuncType(astType Type) (funcType Type) {
	switch subAst := astType.(type) {
	case *MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			getType := GetAllFuncType(oneType)
			if getType != nil {
				return getType
			}
		}

		return
	case *FuncType:
		return astType
	}

	return
}

// GetAstTypeLoc 获取AstType的位置信息
func GetAstTypeLoc(astType Type) lexer.Location {
	switch subAst := astType.(type) {
	case *MultiType:
		return subAst.Loc
	case *FuncType:
		return subAst.Loc
	case *NormalType:
		return subAst.NameLoc
	case *ArrayType:
		return subAst.Loc
	case *TableType:
		return subAst.Loc
	}

	return lexer.Location{}
}

//获取部分注解类型的字符串名称 用于类型检查
func GetAstTypeName(astType Type) string {
	switch subAst := astType.(type) {

	case *NormalType:
		return subAst.StrName
	case *TableType:
		return "table"
	}

	return "any"
}

// GetTypeColorLocVec 获取所有Type产生的着色位置信息
func GetTypeColorLocVec(astType Type) (locVec []lexer.Location) {
	switch subAst := astType.(type) {
	case *MultiType:
		for _, oneType := range subAst.TypeList {
			subLocVec := GetTypeColorLocVec(oneType)
			locVec = append(locVec, subLocVec...)
		}
	case *FuncType:
		// fun key location
		locVec = append(locVec, subAst.FunLoc)

		// 获取所有的参数类型位置信息
		for _, oneType := range subAst.ParamTypeList {
			subLocVec := GetTypeColorLocVec(oneType)
			locVec = append(locVec, subLocVec...)
		}

		// 获取所有的函数返回类型位置信息
		for _, oneType := range subAst.ReturnTypeList {
			subLocVec := GetTypeColorLocVec(oneType)
			locVec = append(locVec, subLocVec...)
		}
	case *NormalType:
		if subAst.ShowColor {
			locVec = append(locVec, subAst.NameLoc)
		}
	case *ArrayType:
		locVec = GetTypeColorLocVec(subAst.ItemType)
	case *TableType:
		locVec = append(locVec, subAst.TableStrLoc)

		if !subAst.EmptyFlag {
			keyLocVec := GetTypeColorLocVec(subAst.KeyType)
			locVec = append(locVec, keyLocVec...)

			valueLocVec := GetTypeColorLocVec(subAst.ValueType)
			locVec = append(locVec, valueLocVec...)
		}
	}

	return locVec
}

func colInLocation(loc lexer.Location, col int) bool {
	if col >= loc.StartColumn && col <= loc.EndColumn {
		return true
	}

	return false
}

// GetTypeLocInfo 获取类型指定的列信息
func GetTypeLocInfo(astType Type, col int) (typeStr string, noticeStr string) {
	switch subAst := astType.(type) {
	case *MultiType:
		for _, oneType := range subAst.TypeList {
			typeStr, noticeStr = GetTypeLocInfo(oneType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr
			}
		}
	case *FuncType:
		for _, oneParamLoc := range subAst.ParamNameLocList {
			if colInLocation(oneParamLoc, col) {
				typeStr = ""
				noticeStr = "function param"
				return
			}
		}

		// 获取所有的参数类型位置信息
		for _, oneType := range subAst.ParamTypeList {
			typeStr, noticeStr = GetTypeLocInfo(oneType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr
			}
		}

		// 获取所有的函数返回类型位置信息
		for _, oneType := range subAst.ReturnTypeList {
			typeStr, noticeStr = GetTypeLocInfo(oneType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr
			}
		}
	case *NormalType:
		if colInLocation(subAst.NameLoc, col) {
			typeStr = subAst.StrName
			noticeStr = ""
			return typeStr, noticeStr
		}
	case *ArrayType:
		typeStr, noticeStr = GetTypeLocInfo(subAst.ItemType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr
		}
	case *TableType:
		if !subAst.EmptyFlag {
			typeStr, noticeStr = GetTypeLocInfo(subAst.KeyType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr
			}

			typeStr, noticeStr = GetTypeLocInfo(subAst.ValueType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr
			}
		}
	}

	return
}

// GetStateLocInfo 获取一个State的指定列的信息
func GetStateLocInfo(oneState AnnotateState, col int) (typeStr string, noticeStr string, commentStr string) {
	switch state := oneState.(type) {
	case *AnnotateAliasState:
		if colInLocation(state.NameLoc, col) {
			typeStr = ""
			noticeStr = "alias name"
			commentStr = state.Comment
			return
		}

		typeStr, noticeStr = GetTypeLocInfo(state.AliasType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr, ""
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}

	case *AnnotateClassState:
		if colInLocation(state.NameLoc, col) {
			typeStr = ""
			noticeStr = "class name"
			commentStr = state.Comment
			return
		}

		for index, oneLoc := range state.ParentLocList {
			if colInLocation(oneLoc, col) {
				typeStr = state.ParentNameList[index]
				noticeStr = ""
				return
			}
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}

	case *AnnotateFieldState:
		if colInLocation(state.NameLoc, col) {
			typeStr = ""
			noticeStr = "field name"
			return
		}

		typeStr, noticeStr = GetTypeLocInfo(state.FiledType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr, ""
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}

	case *AnnotateParamState:
		if colInLocation(state.NameLoc, col) {
			typeStr = ""
			noticeStr = "param name"
			return
		}

		typeStr, noticeStr = GetTypeLocInfo(state.ParamType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr, ""
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}

	case *AnnotateTypeState:
		for _, subType := range state.ListType {
			typeStr, noticeStr = GetTypeLocInfo(subType, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr, ""
			}
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}
	case *AnnotateReturnState:
		for _, subReturn := range state.ReturnTypeList {
			typeStr, noticeStr = GetTypeLocInfo(subReturn, col)
			if typeStr != "" || noticeStr != "" {
				return typeStr, noticeStr, ""
			}
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}
	case *AnnotateGenericState:
		for _, oneLoc := range state.NameLocList {
			if colInLocation(oneLoc, col) {
				typeStr = ""
				noticeStr = "generic name"
				return
			}
		}

		for index, oneLoc := range state.ParentLocList {
			if colInLocation(oneLoc, col) {
				typeStr = state.ParentNameList[index]
				noticeStr = ""
				return
			}
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}
	case *AnnotateOverloadState:
		typeStr, noticeStr = GetTypeLocInfo(state.OverFunType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr, ""
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}

	case *AnnotateVarargState:
		typeStr, noticeStr = GetTypeLocInfo(state.VarargType, col)
		if typeStr != "" || noticeStr != "" {
			return typeStr, noticeStr, ""
		}

		if colInLocation(state.CommentLoc, col) {
			typeStr = ""
			noticeStr = "comment info"
			commentStr = state.Comment
			return
		}
	}

	return
}

// GetAllStrAndLocList 获取Type的所有可能的字符串，及位置信息
func GetAllStrAndLocList(astType Type) (strList []string, locList []lexer.Location) {
	switch subAst := astType.(type) {
	case *MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			tmpStrList, tmpLocList := GetAllStrAndLocList(oneType)
			strList = append(strList, tmpStrList...)
			locList = append(locList, tmpLocList...)
		}

		return strList, locList
	case *NormalType:
		strList = append(strList, subAst.StrName)
		locList = append(locList, subAst.NameLoc)
		return strList, locList

	case *ArrayType:
		return GetAllStrAndLocList(subAst.ItemType)

	case *TableType:
		if subAst.EmptyFlag {
			return
		}

		tmpStrList, tmpLocList := GetAllStrAndLocList(subAst.KeyType)
		strList = append(strList, tmpStrList...)
		locList = append(locList, tmpLocList...)

		tmpStrList, tmpLocList = GetAllStrAndLocList(subAst.ValueType)
		strList = append(strList, tmpStrList...)
		locList = append(locList, tmpLocList...)

		return
	case *FuncType:
		for _, oneType := range subAst.ParamTypeList {
			tmpStrList, tmpLocList := GetAllStrAndLocList(oneType)
			strList = append(strList, tmpStrList...)
			locList = append(locList, tmpLocList...)
		}

		for _, oneType := range subAst.ReturnTypeList {
			tmpStrList, tmpLocList := GetAllStrAndLocList(oneType)
			strList = append(strList, tmpStrList...)
			locList = append(locList, tmpLocList...)
		}

		return
	}

	return
}
